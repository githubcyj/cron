package worker

import (
	"crontab/common"
	"fmt"
	"gopkg.in/mgo.v2"
	"log"
	"time"
)

//mongo日志存储
type LogMgr struct {
	client         *mgo.Database
	collection     *mgo.Collection
	logChan        chan *common.JobLog
	autoCommitChan chan *common.LogBatch //在这里面的需要立即提交
}

var (
	GLogMgr *LogMgr
)

func InitLogMgr() (err error) {
	var (
		session    *mgo.Session
		client     *mgo.Database
		collection *mgo.Collection
	)
	//建立连接
	if session, err = mgo.Dial("127.0.0.1:27017"); err != nil {
		fmt.Println(err)
		return
	}

	client = session.DB("cron")
	collection = client.C("log")
	GLogMgr = &LogMgr{
		client:         client,
		collection:     collection,
		logChan:        make(chan *common.JobLog, GConfig.AutoCommitCount),
		autoCommitChan: make(chan *common.LogBatch, GConfig.AutoCommitCount),
	}
	go GLogMgr.writeLogLoop()
	return
}

//如果日志队列满则马上提交，如果超过超时时间也马上提交
func (logMgr *LogMgr) writeLogLoop() {

	var (
		log          *common.JobLog
		logBatch     *common.LogBatch
		timeoutBatch *common.LogBatch
		timer        *time.Timer
	)

	for {
		select {
		case log = <-logMgr.logChan: //读取日志队列
			if logBatch == nil {
				logBatch = &common.LogBatch{}
				//超时自动提交
				timer = time.AfterFunc(time.Duration(GConfig.AutoCommitTime)*time.Millisecond,
					func(batch *common.LogBatch) func() {
						return func() {
							logMgr.autoCommitChan <- batch //写入自动提交队列
						}

					}(logBatch))
			}

			//把新的日志加入到批次中
			logBatch.Logs = append(logBatch.Logs, log)

			//logMgr.WriteLog("日志写入批次,log:" + log.JobName)

			//批次满了
			if len(logBatch.Logs) >= GConfig.AutoCommitCount {
				logMgr.saveLogs(logBatch)
				//清空logBatch
				logBatch = nil
			}

		case timeoutBatch = <-logMgr.autoCommitChan:
			//判断超时批次是否是当前批次，如果是，则跳过当前批次
			if timeoutBatch != logBatch {
				continue
			}
			logMgr.saveLogs(timeoutBatch)
			//清空logBatch
			logBatch = nil
			//取消定时器
			timer.Stop()
			//logMgr.WriteLog("日志提交成功")
		}
	}
}

//向文件中写日志
func (logMgr *LogMgr) saveLogs(batch *common.LogBatch) {
	var (
		err error
	)
	if err = logMgr.collection.Insert(batch.Logs...); err != nil {
		logMgr.WriteLog("写入数据库出错,err：" + err.Error())
	}
}

//存储日志到队列中
func (logMgr *LogMgr) Append(log *common.JobLog) {
	logMgr.logChan <- log
}

func (logMgr *LogMgr) WriteLog(msg string) {

	log.Println(msg)
}
