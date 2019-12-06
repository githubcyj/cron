package common

import (
	"context"
	"crontab/model"
	"fmt"
	"github.com/gorhill/cronexpr"
	"strings"
	"time"
)

//http应答
type HttpReply struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

//任务事件
type JobEvent struct {
	Pipeline *model.Pipeline //任务内容
	Event    int             //任务事件
	Type     int             //任务类型
}

type JobExecuteInfo struct {
	Pipeline   *model.Pipeline
	CancelCtx  context.Context    //任务command的context
	CancelFunc context.CancelFunc //用于取消任务执行的cancel函数
}

//任务调度计划
type JobSchedulePlan struct {
	Pipeline *model.Pipeline      //要调度的任务
	Expr     *cronexpr.Expression //解析好的执行计划
	NextTime time.Time            //下次执行时间
}

//延时任务调度计划
type JobScheduleTimer struct {
	Pipeline *model.Pipeline      //要调度的任务
	Expr     *cronexpr.Expression //解析好的执行计划
}

type JobExecuteResult struct {
	JobRecord     []*model.JobRecord
	PiplineRecord *model.PipelineRecord
}

//日志信息
type JobLog struct {
	JobName      string `json:"jobName" bson:"jobName"`           //任务名字
	Command      string `json:"command" bson:"command"`           //脚本命令
	Err          string `json:"err" bson:"err"`                   //错误原因
	Output       string `json:"output" bson:"output"`             //脚本输出
	PlanTime     int64  `json:"planTime" bson:"planTime"`         //计划开始时间
	SchedultTime int64  `json:"scheduleTime" bson:"scheduleTime"` //实际调度时间
	StartTime    int64  `json:"startTime" bson:"startTime"`       // 任务执行开始时间
	EndTime      int64  `json:"endTime" bson:"endTime"`           //任务执行结束时间
	Runtime      int64  `json:"runtime" bson:"runtime"`           //任务运行时间
}

type LogBatch struct {
	Logs []interface{}
}

func BuildJobSchedulePlan(pipeline *model.Pipeline) (jobSchedulePlan *JobSchedulePlan, err error) {

	var (
		cronExpr *cronexpr.Expression
	)

	if cronExpr, err = cronexpr.Parse(pipeline.CronExpr); err != nil {
		return
	}

	jobSchedulePlan = &JobSchedulePlan{
		Pipeline: pipeline,
		Expr:     cronExpr,
		NextTime: cronExpr.Next(time.Now()),
	}

	return
}

//
//func BuildJobScheduleTimer(job *Job) (jobSchedulePlan *JobScheduleTimer, err error) {
//
//	var (
//		cronExpr *cronexpr.Expression
//	)
//
//	if cronExpr, err = cronexpr.Parse(job.CronExpr); err != nil {
//		return
//	}
//
//	jobSchedulePlan = &JobScheduleTimer{
//		Job:  job,
//		Expr: cronExpr,
//	}
//
//	return
//}
//
//func BuildJobTimerExecuteInfo(jobSchedulePlan *JobScheduleTimer) (jobExecuteInfo *JobExecuteInfo) {
//
//	jobExecuteInfo = &JobExecuteInfo{
//		Job:          jobSchedulePlan.Job,
//		PlanTime:     time.Now(),
//		ExcutingTime: time.Now(),
//	}
//
//	return
//}

func BuildJobExecuteInfo(jobSchedulePlan *JobSchedulePlan) (jobExecuteInfo *JobExecuteInfo) {

	jobExecuteInfo = &JobExecuteInfo{
		Pipeline: jobSchedulePlan.Pipeline,
	}

	return
}

//构建任务事件
func BuildJobEvent(job *Job, event int) (jobEvent *JobEvent) {

	jobEvent = &JobEvent{
		Job:     job,
		Event:   event,
		JobType: job.Type,
	}

	return
}

//构建http应答
func BuildResponse(errno int, msg string, data interface{}) HttpReply {
	var (
		resp HttpReply
	)
	resp.Errno = errno
	resp.Msg = msg
	resp.Data = data
	//序列化
	return resp
}

func ExtracJobName(name string) string {
	return strings.TrimPrefix(name, SAVE_JOB_DIR)
}

func ExtracWorkIp(name string) string {
	return strings.TrimPrefix(name, JOB_WORKER_DIR)
}

func ExtracKillJob(name string) string {
	return strings.TrimPrefix(name, KILL_JOB_DIR)
}
