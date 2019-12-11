package worker

import (
	"context"
	"github.com/crontab/common"
	"github.com/crontab/constants"
	"go.etcd.io/etcd/clientv3"
)

type JobLock struct {
	Kv         clientv3.KV
	Lease      clientv3.Lease
	Txn        clientv3.Txn
	JobName    string
	LeaseId    clientv3.LeaseID
	cancelCtx  context.Context
	cancelFunc context.CancelFunc
	isLocked   bool
}

var (
	GJobLock *JobLock
)

func InitJobLock(kv clientv3.KV, lease clientv3.Lease, jobName string) {
	GJobLock = &JobLock{
		Kv:      kv,
		Lease:   lease,
		JobName: jobName,
	}
	return
}

//尝试上锁
func (jobLock *JobLock) TryLock() (err error) {

	var (
		leaseResp         *clientv3.LeaseGrantResponse
		leaseId           clientv3.LeaseID
		cancelCtx         context.Context
		cancelFunc        context.CancelFunc
		leaseKeepRespChan <-chan *clientv3.LeaseKeepAliveResponse
		txn               clientv3.Txn
		lockKey           string
		txnResp           *clientv3.TxnResponse
	)

	//创建一个5秒的租约
	if leaseResp, err = jobLock.Lease.Grant(context.TODO(), 5); err != nil {
		goto FAIL
	}

	//创建一个用于取消自动续租的context
	cancelCtx, cancelFunc = context.WithCancel(context.TODO())

	//租约id
	leaseId = leaseResp.ID

	//自动续租
	if leaseKeepRespChan, err = jobLock.Lease.KeepAlive(cancelCtx, leaseId); err != nil {
		goto FAIL
	}

	//启动一个协程来消费续租应答
	go func() {
		var (
			leasekeepResp *clientv3.LeaseKeepAliveResponse
		)
		for {
			select {
			case leasekeepResp = <-leaseKeepRespChan:
				if leasekeepResp == nil { //自动租约被取消
					goto END
				}
			}
		}
	END:
	}()

	//创建事务
	txn = jobLock.Kv.Txn(context.TODO())

	//锁路径
	lockKey = constants.JOB_LOCK_DIR + jobLock.JobName

	//事务抢锁
	//判断是否已经上锁
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet(lockKey))

	//提交事务
	if txnResp, err = txn.Commit(); err != nil {
		goto FAIL
	}

	//成功返回，失败释放租约
	if !txnResp.Succeeded {
		err = common.ERR_LOCK_ALREADY_REQUIRED
		goto FAIL
	}

	//抢锁成功
	jobLock.LeaseId = leaseId
	jobLock.cancelCtx = cancelCtx
	jobLock.cancelFunc = cancelFunc
	jobLock.isLocked = true

	//GLogMgr.WriteLog("上锁成功：" + lockKey)

	return

FAIL:
	cancelFunc()                                  //取消自动续租
	jobLock.Lease.Revoke(context.TODO(), leaseId) //释放租约
	return
}

//释放锁
func (jobLock *JobLock) Unlock() {
	if jobLock.isLocked {
		//GLogMgr.WriteLog("释放锁：" + jobLock.JobName)
		jobLock.cancelFunc()
		jobLock.Lease.Revoke(context.TODO(), jobLock.LeaseId)
	}
}
