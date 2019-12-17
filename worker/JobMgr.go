package worker

import (
	"context"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/crontab/common"
	"github.com/crontab/constants"
	"github.com/crontab/model"
	"github.com/crontab/util"
	"github.com/wxnacy/wgo/arrays"
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

	//go GJobMgr.WatchDel()

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
		jobEvent        *common.JobEvent
		pipeline        *model.Pipeline
		local           string
		index           int
	)

	jobKey = constants.SAVE_JOB_DIR
	local, _ = util.GetLocalIp() //本机ip
	//创建一个wacher
	watcher = jobMgr.watcher

	if getResp, err = jobMgr.Kv.Get(context.TODO(), jobKey, clientv3.WithPrefix()); err != nil {
		return
	}

	GLogMgr.WriteLog("从etcd中读取到的任务总数：" + strconv.Itoa(len(getResp.Kvs)))
	//如果etcd中有任务，则需要通知调度协程去调度
	for _, keyPair = range getResp.Kvs {
		pipeline = &model.Pipeline{}
		//反序列化
		if err = json.Unmarshal(keyPair.Value, pipeline); err != nil {
			GLogMgr.WriteLog("解析etcd中的任务出错：" + err.Error())
		}
		//判断流水线是否绑定节点
		if len(pipeline.Nodes) != 0 { //流水线绑定了节点
			//判断流水线中的节点是否有本机节点
			if index = arrays.Contains(pipeline.Nodes, local); index == -1 { //流水线不在本机节点，则在本机不执行
				GLogMgr.WriteLog("流水线不在此节点执行")
				return
			}
		}
		GLogMgr.WriteLog("从etcd读取到的任务：" + pipeline.Name)
		//创建jobEvent
		jobEvent = common.BuildJobEvent(pipeline, constants.SAVE_JOB_EVENT)
		//通知调度协程
		GScheduler.PushScheduler(jobEvent)
	}

	//设置一个协程来消费watcherChan
	go func() {
		var (
			watchResp clientv3.WatchResponse
			event     *clientv3.Event
			jobEvent  *common.JobEvent
		)
		watcherRevision = getResp.Header.Revision + 1

		watcherChan = watcher.Watch(context.TODO(), jobKey, clientv3.WithPrefix(), clientv3.WithRev(watcherRevision), clientv3.WithPrevKV())
		for watchResp = range watcherChan { //处理kv变化,需要从发布任务之后的版本开始监听
			for _, event = range watchResp.Events {
				switch event.Type {
				case mvccpb.PUT: //如果是put操作
					GLogMgr.WriteLog("添加任务，任务：" + string(event.Kv.Value))
					//反序列化
					pipeline = &model.Pipeline{}
					if err = json.Unmarshal(event.Kv.Value, pipeline); err != nil {
						GLogMgr.WriteLog("解析流水线出错：" + err.Error())
						return
					}
					//判断流水线是否绑定节点
					if len(pipeline.Nodes) != 0 { //流水线绑定了节点
						//判断流水线中的节点是否有本机节点
						if index = arrays.Contains(pipeline.Nodes, local); index == -1 { //流水线不在本机节点，则在本机不执行
							GLogMgr.WriteLog("流水线不在此节点执行")
							return
						}
					}
					//判断任务类型
					//构建任务事件
					jobEvent = common.BuildJobEvent(pipeline, constants.SAVE_JOB_EVENT)
				case mvccpb.DELETE: //如果是del操作
					//GLogMgr.WriteLog("删除任务，任务：" + string(event.Kv.Key))

					////需要从/cron/job/save/job2中提取出job2
					//jobName = common.ExtracJobName(string(event.Kv.Key))
					//job = &common.Job{
					//	Name: jobName,
					//}
					////构建任务事件
					//jobEvent = common.BuildJobEvent(job, constants.DELETE_JOB_EVENT)
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
		watcher    clientv3.Watcher
		killDir    string
		watchChan  clientv3.WatchChan
		watchResp  clientv3.WatchResponse
		event      *clientv3.Event
		jobEvent   *common.JobEvent
		pipeliine  *model.Pipeline
		pipelineId string
	)

	//创建一个监听者
	watcher = jobMgr.watcher

	//强杀任务目录
	killDir = constants.KILL_JOB_DIR

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
					pipelineId = string(event.Kv.Value)
					pipeliine = &model.Pipeline{PipelineId: pipelineId}
					//构建任务事件
					jobEvent = common.BuildJobEvent(pipeliine, constants.KILL_JOB_EVENT)
					//通知调度器
					GScheduler.PushScheduler(jobEvent)
				}
			}
		}
	}()

}

//
////监听删除任务
//func (jobMgr *JobMgr) WatchDel() {
//	var (
//		watcher   clientv3.Watcher
//		delDir    string
//		watchChan clientv3.WatchChan
//		watchResp clientv3.WatchResponse
//		event     *clientv3.Event
//		jobId     string
//		jobEvent  *common.JobEvent
//		err       error
//	)
//
//	//创建一个监听者
//	watcher = jobMgr.watcher
//
//	//强杀任务目录
//	delDir = constants.JOB_DEl_DIR
//
//	//创建一个协程来消费
//	go func() {
//		//监听该目录后续变化
//		watchChan = watcher.Watch(context.TODO(), delDir, clientv3.WithPrefix())
//		for watchResp = range watchChan {
//			for _, event = range watchResp.Events {
//				switch event.Type {
//				//强杀任务
//				case mvccpb.PUT: // 杀死某个任务的事件
//					GLogMgr.WriteLog("监听到删除任务")
//					//获得对应的任务名
//					jobId = string(event.Kv.Value)
//					//从redis中获取job
//					if job, err = GRedis.GetSingleJob(jobId); err != nil {
//						GLogMgr.WriteLog("从redis中查找任务失败，任务id:" + jobId)
//					}
//					//构建任务事件
//					jobEvent = common.BuildJobEvent(job, constants.DELETE_JOB_EVENT)
//					//通知调度器
//					GScheduler.PushScheduler(jobEvent)
//				}
//			}
//		}
//	}()
//}
