package manager

import (
	"context"
	"crontab/constants"
	"crontab/master"
	"go.etcd.io/etcd/clientv3"
	"time"
)

//etcd客户端
type JobMgr struct {
	Client *clientv3.Client
	Kv     clientv3.KV
	Lease  clientv3.Lease
}

var (
	GJobMgr *JobMgr
)

//创建etcd连接
func InitJobManager() (err error) {
	var (
		client *clientv3.Client
		config clientv3.Config
		kv     clientv3.KV
		lease  clientv3.Lease
		jobMgr *JobMgr
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

	jobMgr = &JobMgr{
		Client: client,
		Kv:     kv,
		Lease:  lease,
	}

	//赋值单例
	GJobMgr = jobMgr

	return
}

//恢复逻辑删除的etcd
func (jobMgr *JobMgr) RecoverJob(jobIds []string) (err error) {
	var (
		deleteKey string
		jobId     string
		saveKey   string
	)
	for _, jobId = range jobIds {
		deleteKey = constants.JOB_DEl_DIR + jobId

		if _, err = jobMgr.Kv.Delete(context.TODO(), deleteKey); err != nil {
			return
		}

		saveKey = constants.SAVE_JOB_DIR + jobId

		//重新保存对应的etcd
		if _, err = jobMgr.Kv.Put(context.TODO(), saveKey, jobId); err != nil {
			return
		}
	}
	return
}

//先删除后保存
func (jobMgr *JobMgr) DeleteJobForLogic(jobIds []string) (err error) {
	var (
		deleteKey string
		oldKey    string
		jobId     string
	)
	for _, jobId = range jobIds {
		oldKey = constants.SAVE_JOB_DIR + jobId

		if _, err = jobMgr.Kv.Delete(context.TODO(), oldKey); err != nil {
			return
		}

		deleteKey = constants.JOB_DEl_DIR + jobId
		//重新保存对应的etcd
		if _, err = jobMgr.Kv.Put(context.TODO(), deleteKey, jobId); err != nil {
			return
		}
	}
	return
}

//
//func (jobMgr *JobMgr) SaveJob(job *common.Job) (oldJob *common.Job, err error) {
//	var (
//		saveKey     string
//		putResponse *clientv3.PutResponse
//		old         common.Job
//	)
//
//	//保存的路径
//	saveKey = constants.SAVE_JOB_DIR + job.JobId
//
//	//需要传回上一次保存的数据，用以反馈
//	if putResponse, err = GJobMgr.Kv.Put(context.TODO(), saveKey, job.JobId, clientv3.WithPrevKV()); err != nil {
//		return
//	}
//	GLogMgr.WriteLog("添加任务成功,任务：" + job.String())
//
//	//如果是更新，则需要返回旧值
//	if putResponse.PrevKv != nil {
//		//反序列化
//		if err = json.Unmarshal(putResponse.PrevKv.Value, &old); err != nil {
//			err = nil //已经保存成功后，不需要关心旧值的内容
//			return
//		}
//
//		oldJob = &old
//	} else {
//		oldJob = job
//	}
//
//	return
//}
//
////主要是为了让watch监听到一次put操作
//func (jobMgr *JobMgr) UpdateJob(job *common.Job) (err error) {
//	var (
//		deleteKey string
//	)
//	deleteKey = constants.SAVE_JOB_DIR + job.JobId
//
//	//重新保存对应的etcd
//	if _, err = jobMgr.Kv.Put(context.TODO(), deleteKey, job.JobId); err != nil {
//		return
//	}
//	return
//}
//
//func (jobMgr *JobMgr) DelJob(jobIds []string) (ids []string, err error) {
//	var (
//		delKey  string
//		delResp *clientv3.DeleteResponse
//		id      string
//	)
//
//	ids = make([]string, 0)
//
//	for _, id = range jobIds {
//		delKey = constants.JOB_DEl_DIR + id
//		if delResp, err = jobMgr.Kv.Delete(context.TODO(), delKey, clientv3.WithPrevKV()); err != nil {
//			return ids, err
//		}
//		//返回被删除的信息
//		if len(delResp.PrevKvs) > 0 {
//			ids = append(ids, string(delResp.PrevKvs[0].Value))
//		}
//		GLogMgr.WriteLog("删除任务成功,任务：" + string(delResp.PrevKvs[0].Value))
//	}
//	return ids, nil
//}
//
//func (jobMgr *JobMgr) ListJob() (jobArr []*common.Job, err error) {
//	var (
//		listKey string
//		getResp *clientv3.GetResponse
//		value   *mvccpb.KeyValue
//		job     *common.Job
//	)
//
//	listKey = constants.SAVE_JOB_DIR
//
//	if getResp, err = jobMgr.Kv.Get(context.TODO(), listKey, clientv3.WithPrefix()); err != nil {
//		return
//	}
//
//	//初始化jobArr，可以使用len(jobArr)来判断是否有值
//	jobArr = make([]*common.Job, 0)
//
//	//遍历kv
//	for _, value = range getResp.Kvs {
//		job = &common.Job{}
//		//反序列化
//		if err = json.Unmarshal(value.Value, job); err != nil {
//			return
//		}
//		jobArr = append(jobArr, job)
//	}
//	return
//}
//
//func (jobMgr *JobMgr) KillJob(name string) (oldJob common.Job, err error) {
//
//	var (
//		killKey        string
//		delKey         string
//		putResp        *clientv3.PutResponse
//		leaseGrantResp *clientv3.LeaseGrantResponse
//		leaseId        clientv3.LeaseID
//		//delResp        *clientv3.DeleteResponse
//	)
//
//	killKey = constants.KILL_JOB_DIR + name
//	delKey = constants.SAVE_JOB_DIR + name
//
//	//让workder监听到一次put操作，创建一个租约让其稍后自动过期
//	if leaseGrantResp, err = jobMgr.Lease.Grant(context.TODO(), 1); err != nil {
//		return
//	}
//
//	leaseId = leaseGrantResp.ID
//
//	if putResp, err = jobMgr.Kv.Put(context.TODO(), killKey, "", clientv3.WithLease(leaseId)); err != nil {
//		return
//	}
//
//	if _, err = jobMgr.Kv.Delete(context.TODO(), delKey); err != nil {
//		fmt.Println(err.Error())
//		return
//	}
//
//	if putResp.PrevKv != nil {
//		//反序列化json
//		return
//	}
//	return
//}
