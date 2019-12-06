package worker

import (
	"crontab/common"
	"strconv"
	"time"
)

//调度器，用来调度任务
type Scheduler struct {
	JobEventChan           chan *common.JobEvent               //etcd任务事件队列
	JobPlanTable           map[string]*common.JobSchedulePlan  //任务调度计划表
	JobExecutingTable      map[string]*common.JobExecuteInfo   //任务调度执行表
	JobResultChan          chan *common.JobExecuteResult       //任务结果队列
	JobTimerTable          map[string]*common.JobScheduleTimer //延时任务调度计划表
	JobTimerExecutingTable map[string]*common.JobExecuteInfo   //任务调度执行表
}

var (
	GScheduler *Scheduler
)

func InitScheduler() (err error) {
	GScheduler = &Scheduler{
		JobEventChan:      make(chan *common.JobEvent, 1000),
		JobPlanTable:      make(map[string]*common.JobSchedulePlan),
		JobExecutingTable: make(map[string]*common.JobExecuteInfo),
		JobResultChan:     make(chan *common.JobExecuteResult, 1000),
	}

	//启动调度协程
	go GScheduler.SchedulerLoop()
	return
}

func (scheduler *Scheduler) SchedulerLoop() {
	var (
		jobEvent   *common.JobEvent
		jobResult  *common.JobExecuteResult
		timerAfter time.Duration
		timer      *time.Timer //这里设置定时器，是为了让之后的任务准时执行
		log        *common.JobLog
	)

	timerAfter = GScheduler.TrySchedule()
	//先随机睡眠1秒
	timer = time.NewTimer(timerAfter)
	//保持监听
	for {
		select {
		case jobEvent = <-scheduler.JobEventChan: //事件任务队列
			//对任务事件进行增删改查
			GLogMgr.WriteLog("从任务事件中读取任务：" + jobEvent.Job.String())
			scheduler.OperateJobEvent(jobEvent)
		case <-timer.C:
		case jobResult = <-scheduler.JobResultChan: //监听结果任务队列
			if jobResult.JobExecuteInfo.Job.Type == common.CRON_JOB_TYPE {
				//任务执行完毕，需要从任务执行队列中删除任务
				GLogMgr.WriteLog("任务执行完毕，从任务执行表中删除，任务：" + jobResult.JobExecuteInfo.Job.String())
				delete(scheduler.JobExecutingTable, jobResult.JobExecuteInfo.Job.Name)
				//GLogMgr.WriteLog("任务执行表中任务个数：" + strconv.Itoa(len(scheduler.JobExecutingTable)))
				for jobExecuteInfo, _ := range scheduler.JobExecutingTable {
					GLogMgr.WriteLog("任务执行表中的任务：" + jobExecuteInfo)
				}
			}
			if jobResult.JobExecuteInfo.Job.Type == common.DELAY_JOB_TYPE {
				GLogMgr.WriteLog("任务执行完毕，从任务执行表中删除，任务：" + jobResult.JobExecuteInfo.Job.String())
				delete(scheduler.JobTimerExecutingTable, jobResult.JobExecuteInfo.Job.Name)
			}
			//将结果保存到日志
			go func() {
				log = &common.JobLog{
					JobName:      jobResult.JobExecuteInfo.Job.Name,
					Command:      jobResult.JobExecuteInfo.Job.Command,
					Output:       string(jobResult.Output),
					PlanTime:     jobResult.JobExecuteInfo.PlanTime.Unix(),
					SchedultTime: jobResult.JobExecuteInfo.ExcutingTime.Unix(),
					StartTime:    jobResult.StartTime.Unix(),
					EndTime:      jobResult.EndTime.Unix(),
				}

				if jobResult.Err == nil {
					log.Err = ""
				} else {
					log.Err = jobResult.Err.Error()
				}

				GLogMgr.Append(log)
			}()
		}
		timerAfter = GScheduler.TrySchedule()
		timer.Reset(timerAfter)
		GScheduler.ScheduleTime()
	}
}

//调度延时消息，延时消息只需要调度一次
func (scheduler *Scheduler) ScheduleTime() {

	var (
		jobScheduleTimer *common.JobScheduleTimer
	)

	if len(scheduler.JobTimerTable) == 0 {
		return
	}
	for _, jobScheduleTimer = range scheduler.JobTimerTable {
		GLogMgr.WriteLog("开始调度延时任务")
		GScheduler.TryStartTimerJob(jobScheduleTimer)
	}
}

//进行任务调度，调度定时消息
func (scheduler *Scheduler) TrySchedule() (timeAfter time.Duration) {

	var (
		jobSchedulePlan *common.JobSchedulePlan
		now             time.Time
		nearTime        *time.Time
	)

	//如果任务表为空，则睡眠
	if len(scheduler.JobPlanTable) == 0 {
		timeAfter = 1 * time.Second
		//GLogMgr.WriteLog("任务表为空")
		return
	}

	now = time.Now()

	//1. 从计划任务表中选择离当前时间最近的且不在执行中的任务执行
	//遍历查找job
	for _, jobSchedulePlan = range GScheduler.JobPlanTable {
		GLogMgr.WriteLog(jobSchedulePlan.Job.Name + "的下次执行时间" + jobSchedulePlan.NextTime.String())
		GLogMgr.WriteLog("当前时间：" + now.String())
		if jobSchedulePlan.NextTime.Before(now) || jobSchedulePlan.NextTime.Equal(now) {
			//尝试执行任务
			GLogMgr.WriteLog("开始执行任务：" + jobSchedulePlan.Job.String())
			scheduler.TryStartJob(jobSchedulePlan)
			//GLogMgr.WriteLog("结束执行任务：" + jobSchedulePlan.Job.String())
			//更新任务的下次执行时间
			jobSchedulePlan.NextTime = jobSchedulePlan.Expr.Next(now)
			//GLogMgr.WriteLog("任务：" + jobSchedulePlan.Job.Name + "的下次执行时间：" + jobSchedulePlan.NextTime.String())
		}

		//统计最近一个要过期的任务时间
		if nearTime == nil {
			nearTime = &jobSchedulePlan.NextTime
		}
	}

	//下次执行任务的时间间隔，下次执行任务时间-当前时间
	timeAfter = (*nearTime).Sub(now)
	//GLogMgr.WriteLog("任务下次执行的时间间隔：" + timeAfter.String())

	return
}

//尝试执行延时任务
func (scheduler *Scheduler) TryStartTimerJob(jobScheduleTimer *common.JobScheduleTimer) {
	var (
		jobExecuting   bool
		jobExecuteInfo *common.JobExecuteInfo
	)

	if jobExecuteInfo, jobExecuting = scheduler.JobExecutingTable[jobScheduleTimer.Job.Name]; jobExecuting {
		return
	}

	//构建任务执行信息
	jobExecuteInfo = common.BuildJobTimerExecuteInfo(jobScheduleTimer)

	//添加到任务执行表
	scheduler.JobTimerExecutingTable[jobExecuteInfo.Job.Name] = jobExecuteInfo

	//4. 执行任务
	go GExecutor.ExecutingJob(jobExecuteInfo)
}

//尝试执行任务
func (scheduler *Scheduler) TryStartJob(jobSchedulePlan *common.JobSchedulePlan) {
	var (
		jobExecuting   bool
		jobExecuteInfo *common.JobExecuteInfo
		runtime        time.Duration
		realRuntime    time.Duration
	)

	//1. 判断任务是否在执行中，如果任务在执行，则判断任务是否超过运行时间
	if jobExecuteInfo, jobExecuting = scheduler.JobExecutingTable[jobSchedulePlan.Job.Name]; jobExecuting {
		//GLogMgr.WriteLog("判断任务是否超时：" + jobExecuteInfo.Job.Name)
		runtime = time.Now().Sub(jobExecuteInfo.ExcutingTime)
		realRuntime = time.Duration(GConfig.JobRuntime) * time.Millisecond
		GLogMgr.WriteLog(jobExecuteInfo.Job.Name + "执行时间：" + runtime.String())
		GLogMgr.WriteLog("realRuntime:" + realRuntime.String())
		if realRuntime < runtime { //任务运行时间太长，杀死任务并重新调度
			GLogMgr.WriteLog("任务超时，杀死任务：" + jobExecuteInfo.Job.Name)
			jobExecuteInfo.CancelFunc()
			delete(GScheduler.JobExecutingTable, jobExecuteInfo.Job.Name)
		}
		return
	}

	//2. 创建任务执行信息
	jobExecuteInfo = common.BuildJobExecuteInfo(jobSchedulePlan)

	//3. 添加到任务执行表中
	scheduler.JobExecutingTable[jobExecuteInfo.Job.Name] = jobExecuteInfo
	GLogMgr.WriteLog(jobExecuteInfo.Job.Name + "加入任务执行表")

	//4. 执行任务
	go GExecutor.ExecutingJob(jobExecuteInfo)
}

//通知调度器
func (scheduler *Scheduler) PushScheduler(jobEvent *common.JobEvent) {
	//GLogMgr.WriteLog("通知调度器，任务：" + jobEvent.Job.String())
	scheduler.JobEventChan <- jobEvent
}

//通知调度器任务执行完毕
func (scheduler *Scheduler) PushSchedulerResult(info *common.JobExecuteResult) {
	//GLogMgr.WriteLog("通知调度器任务执行完毕：" + info.JobExecuteInfo.Job.String())
	scheduler.JobResultChan <- info
}

func (scheduler *Scheduler) OperateJobEvent(jobEvent *common.JobEvent) (err error) {
	var (
		jobSchedulePlan *common.JobSchedulePlan
		jobExecuteInfo  *common.JobExecuteInfo
		jobExecuting    bool
	)
	//如果是添加任务事件，则向planTable中增加一个事件
	switch jobEvent.Event {
	case common.SAVE_JOB_EVENT:
		if jobEvent.JobType == 0 { //如果是定时任务
			if jobSchedulePlan, err = common.BuildJobSchedulePlan(jobEvent.Job); err != nil {
				GLogMgr.WriteLog("任务加入计划表出错：" + jobEvent.Job.String())
				return
			}
			GLogMgr.WriteLog("任务加入计划表：" + jobEvent.Job.Name)
			scheduler.JobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
		}
		//if jobEvent.JobType == 1 { //如果是延时任务，延时任务只需要执行一次，所以直接加入任务执行表中
		//	jobExecuteInfo = &common.JobExecuteInfo{
		//		Job:          jobEvent.Job,
		//		PlanTime:     jobEvent.Job.TimerExecuter,
		//		ExcutingTime: nil,
		//		CancelCtx:    nil,
		//		CancelFunc:   nil,
		//	}
		//	scheduler.JobTimerTable[jobEvent.Job.JobId] = jobExecuteInfo
		//}
	case common.DELETE_JOB_EVENT:
		GLogMgr.WriteLog("将任务从计划表中删除：" + jobEvent.Job.String())
		delete(GScheduler.JobPlanTable, jobEvent.Job.Name)
	case common.KILL_JOB_EVENT: //强杀任务
		//将任务从计划表中删除
		delete(GScheduler.JobPlanTable, jobEvent.Job.Name)
		GLogMgr.WriteLog("任务" + jobEvent.Job.Name + "被强杀")
		GLogMgr.WriteLog("当前任务计划表中任务数：" + strconv.Itoa(len(GScheduler.JobPlanTable)))
		//取消command执行，判断任务是否在执行
		if jobExecuteInfo, jobExecuting = GScheduler.JobExecutingTable[jobEvent.Job.Name]; jobExecuting {
			jobExecuteInfo.CancelFunc() //取消任务执行
			GLogMgr.WriteLog("任务" + jobExecuteInfo.Job.Name + "被杀死")
		}
	}

	return
}
