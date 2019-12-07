package manager

import (
	"context"
	"crontab/constants"
	"crontab/master"
	"go.etcd.io/etcd/clientv3"
	"time"
)

type MasterMgr struct {
	Kv         clientv3.KV
	Client     *clientv3.Client
	Lease      clientv3.Lease
	IsMaster   bool
	LeaseId    clientv3.LeaseID
	Txn        clientv3.Txn
	CancleCtx  context.Context
	CancelFunc context.CancelFunc
}

var GMasterMgr *MasterMgr

//服务启动去抢占/cron/job/master锁，抢到锁的即为leader
func InitMaster() (err error) {
	var (
		client *clientv3.Client
		kv     clientv3.KV
		config clientv3.Config
		lease  clientv3.Lease
	)

	config = clientv3.Config{
		Endpoints:   master.GConfig.EtcdIP,
		DialTimeout: time.Duration(master.GConfig.EtcdTimeout) * time.Millisecond,
	}

	if client, err = clientv3.New(config); err != nil {
		return
	}

	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	GMasterMgr = &MasterMgr{
		Kv:       kv,
		Client:   client,
		Lease:    lease,
		IsMaster: false,
	}

	go GMasterMgr.voting()
	return
}

func (masterMgr *MasterMgr) voting() {
	var (
		votingKey              string //锁路径
		getResp                *clientv3.GetResponse
		err                    error
		cancleCtx              context.Context
		cancleFunc             context.CancelFunc
		leaseGrantResp         *clientv3.LeaseGrantResponse
		leaseId                clientv3.LeaseID
		leaseKeepAliveRespChan <-chan *clientv3.LeaseKeepAliveResponse
		txn                    clientv3.Txn
		txnResp                *clientv3.TxnResponse
	)

	votingKey = constants.JOB_JOB_MASTER

	if getResp, err = masterMgr.Client.Get(context.TODO(), votingKey, clientv3.WithPrefix()); err != nil {
		return
	}
	cancleCtx, cancleFunc = context.WithCancel(context.TODO())
	//没有master
	if len(getResp.Kvs) == 0 {
		//创建租约
		if leaseGrantResp, err = masterMgr.Lease.Grant(context.TODO(), 5); err != nil {
			return
		}

		leaseId = leaseGrantResp.ID

		//自动续租
		if leaseKeepAliveRespChan, err = masterMgr.Lease.KeepAlive(cancleCtx, leaseId); err != nil {
			return
		}
		//创建一个协程来消费自动续租
		go func() {
			var (
				leaseKeepAliveResp *clientv3.LeaseKeepAliveResponse
			)
			select {
			case leaseKeepAliveResp = <-leaseKeepAliveRespChan:
				if leaseKeepAliveResp == nil { //自动续租被取消
					goto END
				}
			}
		END:
		}()

		//创建事务
		txn = masterMgr.Kv.Txn(context.TODO())

		//事务抢锁
		txn.If(clientv3.Compare(clientv3.CreateRevision(votingKey), "=", 0)).
			Then(clientv3.OpPut(votingKey, "", clientv3.WithLease(leaseId))).
			Else(clientv3.OpGet(votingKey))

		//提交事务
		if txnResp, err = txn.Commit(); err != nil {
			return
		}

		//抢锁失败
		if !txnResp.Succeeded {
			//取消自动续租
			cancleFunc()
			//释放租约
			masterMgr.Lease.Revoke(context.TODO(), leaseId)
			return
		}
		masterMgr.IsMaster = true
		masterMgr.Txn = txn
		masterMgr.LeaseId = leaseId
		masterMgr.CancleCtx = cancleCtx
		masterMgr.CancelFunc = cancleFunc
		return
	}
}

func (masterMgr *MasterMgr) UnVoting() {
	masterMgr.CancelFunc()
	masterMgr.Lease.Revoke(context.TODO(), masterMgr.LeaseId)
}
