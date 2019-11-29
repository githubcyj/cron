package worker

import (
	"context"
	"crontab/common"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"strconv"
	"time"
)

type JobMgr struct {
	Client  *clientv3.Client
	Kv      clientv3.KV
	watcher clientv3.Watcher
}

var (
	GJobMgr *JobMgr
)

func (jobMgr *JobMgr) CreateLock(jobName string) {
	InitJobLock(jobMgr.Kv, jobMgr.Client, jobName)
}

//初始化etcd连接
func InitJobMgr() (err error) {

	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		jobMgr  *JobMgr
		watcher clientv3.Watcher
	)

	config = clientv3.Config{
		Endpoints:   GConfig.EtcdIP,
		DialTimeout: time.Duration(GConfig.EtcdTimeout) * time.Millisecond,
	}

	if client, err = clientv3.New(config); err != nil {
		return
	}

	kv = clientv3.NewKV(client)
	watcher = clientv3.NewWatcher(client)

	jobMgr = &JobMgr{
		Client:  client,
		Kv:      kv,
		watcher: watcher,
	}

	GJobMgr = jobMgr

	//启动任务监听
	if err = GJobMgr.WatchJobs(); err != nil {
		return
	}

	go GJobMgr.watchKill()

	go GJobMgr.WatchDel()

	return
}

//监听任务变化
func (jobMgr *JobMgr) WatchJobs() (err error) {

	var (
		jobKey          string
		watcher         clientv3.Watcher
		watcherChan     clientv3.WatchChan
		getResp         *clientv3.GetResponse
		watcherRevision int64
		keyPair         *mvccpb.KeyValue
		job             *common.Job
		jobEvent        *common.JobEvent
	)

	jobKey = common.SAVE_JOB_DIR

	//创建一个wacher
	watcher = jobMgr.watcher

	if getResp, err = jobMgr.Kv.Get(context.TODO(), jobKey, clientv3.WithPrefix()); err != nil {
		return
	}

	GLogMgr.WriteLog("从etcd中读取到的任务总数：" + strconv.Itoa(len(getResp.Kvs)))
	//如果etcd中有任务，则需要通知调度协程去调度
	for _, keyPair = range getResp.Kvs {
		//从redis中查询任务
		if job, err = GRedis.GetSingleJob(string(keyPair.Value)); err != nil {
			return
		}

		//redis中没有任务，则查询数据库
		if job == nil {
			if job, err = GDB.GetSingleJob(string(keyPair.Value)); err != nil {
				return
			}
			//数据库中没有数据，则直接返回
			if job == nil {
				return
			}
		}

		//GLogMgr.WriteLog("从etcd读取到的任务：" + job.String())
		//创建jobEvent
		jobEvent = common.BuildJobEvent(job, common.SAVE_JOB_EVENT)
		//通知调度协程
		GScheduler.PushScheduler(jobEvent)
	}

	//设置一个协程来消费watcherChan
	go func() {
		var (
			watchResp clientv3.WatchResponse
			event     *clientv3.Event
			job       *common.Job
			jobEvent  *common.JobEvent
			jobName   string
		)
		watcherRevision = getResp.Header.Revision + 1

		watcherChan = watcher.Watch(context.TODO(), jobKey, clientv3.WithPrefix(), clientv3.WithRev(watcherRevision), clientv3.WithPrevKV())
		for watchResp = range watcherChan { //处理kv变化,需要从发布任务之后的版本开始监听
			for _, event = range watchResp.Events {
				switch event.Type {
				case mvccpb.PUT: //如果是put操作
					GLogMgr.WriteLog("添加任务，任务：" + string(event.Kv.Value))
					if job, err = GRedis.GetSingleJob(string(event.Kv.Value)); err != nil {
						return
					}
					//判断任务类型
					//构建任务事件
					jobEvent = common.BuildJobEvent(job, common.SAVE_JOB_EVENT)
				case mvccpb.DELETE: //如果是del操作
					//GLogMgr.WriteLog("删除任务，任务：" + string(event.Kv.Key))

					//需要从/cron/job/save/job2中提取出job2
					jobName = common.ExtracJobName(string(event.Kv.Key))
					job = &common.Job{
						Name: jobName,
					}
					//构建任务事件
					jobEvent = common.BuildJobEvent(job, common.DELETE_JOB_EVENT)
				}
			}

			//通知调度器
			GScheduler.PushScheduler(jobEvent)
		}
	}()

	return
}

//监听强杀任务
func (jobMgr *JobMgr) watchKill() {
	var (
		watcher   clientv3.Watcher
		killDir   string
		watchChan clientv3.WatchChan
		watchResp clientv3.WatchResponse
		event     *clientv3.Event
		jobName   string
		job       *common.Job
		jobEvent  *common.JobEvent
	)

	//创建一个监听者
	watcher = jobMgr.watcher

	//强杀任务目录
	killDir = common.KILL_JOB_DIR

	//创建一个协程来消费
	go func() {
		//监听该目录后续变化
		watchChan = watcher.Watch(context.TODO(), killDir, clientv3.WithPrefix())
		for watchResp = range watchChan {
			for _, event = range watchResp.Events {
				switch event.Type {
				//强杀任务
				case mvccpb.PUT: // 杀死某个任务的事件
					GLogMgr.WriteLog("监听到强杀任务")
					//获得对应的任务名
					jobName = common.ExtracKillJob(string(event.Kv.Key))
					job = &common.Job{
						Name: jobName,
					}
					//构建任务事件
					jobEvent = common.BuildJobEvent(job, common.KILL_JOB_EVENT)
					//通知调度器
					GScheduler.PushScheduler(jobEvent)
				}
			}
		}
	}()

}

//监听删除任务
func (jobMgr *JobMgr) WatchDel() {
	var (
		watcher   clientv3.Watcher
		delDir    string
		watchChan clientv3.WatchChan
		watchResp clientv3.WatchResponse
		event     *clientv3.Event
		jobId     string
		job       *common.Job
		jobEvent  *common.JobEvent
		err       error
	)

	//创建一个监听者
	watcher = jobMgr.watcher

	//强杀任务目录
	delDir = common.JOB_DEl_DIR

	//创建一个协程来消费
	go func() {
		//监听该目录后续变化
		watchChan = watcher.Watch(context.TODO(), delDir, clientv3.WithPrefix())
		for watchResp = range watchChan {
			for _, event = range watchResp.Events {
				switch event.Type {
				//强杀任务
				case mvccpb.PUT: // 杀死某个任务的事件
					GLogMgr.WriteLog("监听到删除任务")
					//获得对应的任务名
					jobId = string(event.Kv.Value)
					//从redis中获取job
					if job, err = GRedis.GetSingleJob(jobId); err != nil {
						GLogMgr.WriteLog("从redis中查找任务失败，任务id:" + jobId)
					}
					//构建任务事件
					jobEvent = common.BuildJobEvent(job, common.DELETE_JOB_EVENT)
					//通知调度器
					GScheduler.PushScheduler(jobEvent)
				}
			}
		}
	}()
}
