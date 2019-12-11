package common

import (
	"context"
	"github.com/crontab/constants"
	"github.com/crontab/model"
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
	Count    int                  //执行次数
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

type LogBatch struct {
	Logs []*JobExecuteResult
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
		Count:    pipeline.RunCount,
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
func BuildJobEvent(pipeline *model.Pipeline, event int) (jobEvent *JobEvent) {

	jobEvent = &JobEvent{
		Pipeline: pipeline,
		Event:    event,
		Type:     pipeline.Type,
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
	return strings.TrimPrefix(name, constants.SAVE_JOB_DIR)
}

func ExtracWorkIp(name string) string {
	return strings.TrimPrefix(name, constants.JOB_WORKER_DIR)
}

func ExtracKillJob(name string) string {
	return strings.TrimPrefix(name, constants.KILL_JOB_DIR)
}
