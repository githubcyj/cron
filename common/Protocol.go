package common

import (
	"context"
	"fmt"
	"github.com/gorhill/cronexpr"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"strings"
	"time"
)

//http应答
type HttpReply struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

//定时任务
type Job struct {
	Id             int       `json:"id"`
	Name           string    `json:"name"`                          //任务名
	Type           int       `json:"type"`                          //任务类型 0：定时任务，1：延时任务
	Command        string    `json:"command"`                       //任务命令
	JobId          string    `json:"jobId"`                         //任务唯一id
	CronExpr       string    `json:"cronExpr"`                      //定时任务：定时任务执行时间
	IsDel          int       `json:"isDel" gorm:"default:'0'`       //0：任务未删除，1：任务已删除
	UpdateCount    int       `json:"updateCount" gorm:"default:'0'` //更新计数
	TimerExecuter  string    `json:"timerExecuter"`                 //延时任务执行时间
	ExecutionCount int       `json:"executionCount"`                //执行次数
	CreateUserId   string    `json:"createUserId"`
	CreateTime     time.Time `json:"createTime"`
	UpdateUserId   string    `json:"updateUserId"`
	UpdateTime     time.Time `json:"updateTime"`
}

func (job *Job) BeforeSave(scope *gorm.Scope) error {
	var (
		err error
	)
	if err = scope.SetColumn("UpdateTime", time.Now()); err != nil {
		return err
	}
	return err
}

func (job *Job) BeforeCreate(scope *gorm.Scope) error {
	var (
		id  uuid.UUID
		err error
		uid string
	)
	id, _ = uuid.NewV4()
	uid = string([]rune(id.String())[:10])
	if err = scope.SetColumn("JobId", uid); err != nil {
		return err
	}

	if err = scope.SetColumn("CreateTime", time.Now()); err != nil {
		return err
	}

	if err = scope.SetColumn("UpdateTime", time.Now()); err != nil {
		return err
	}

	return nil
}

//任务事件
type JobEvent struct {
	Job     *Job //任务内容
	Event   int  //任务事件
	JobType int  //任务类型
}

type JobExecuteInfo struct {
	Job          *Job
	PlanTime     time.Time          //计划执行时间
	ExcutingTime time.Time          //实际执行时间
	CancelCtx    context.Context    //任务command的context
	CancelFunc   context.CancelFunc //用于取消任务执行的cancel函数
}

//任务调度计划
type JobSchedulePlan struct {
	Job      *Job                 //要调度的任务
	Expr     *cronexpr.Expression //解析好的执行计划
	NextTime time.Time            //下次执行时间
}

//延时任务调度计划
type JobScheduleTimer struct {
	Job  *Job                 //要调度的任务
	Expr *cronexpr.Expression //解析好的执行计划
}

type JobExecuteResult struct {
	JobExecuteInfo *JobExecuteInfo //执行状态
	Output         []byte          //脚本输出
	Err            error           //脚本错误原因
	StartTime      time.Time       //启动时间
	EndTime        time.Time       //结束时间
	//ExecutionCount int             //执行次数
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

func BuildJobSchedulePlan(job *Job) (jobSchedulePlan *JobSchedulePlan, err error) {

	var (
		cronExpr *cronexpr.Expression
	)

	if cronExpr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}

	jobSchedulePlan = &JobSchedulePlan{
		Job:      job,
		Expr:     cronExpr,
		NextTime: cronExpr.Next(time.Now()),
	}

	return
}

func BuildJobScheduleTimer(job *Job) (jobSchedulePlan *JobScheduleTimer, err error) {

	var (
		cronExpr *cronexpr.Expression
	)

	if cronExpr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}

	jobSchedulePlan = &JobScheduleTimer{
		Job:  job,
		Expr: cronExpr,
	}

	return
}

func BuildJobTimerExecuteInfo(jobSchedulePlan *JobScheduleTimer) (jobExecuteInfo *JobExecuteInfo) {

	jobExecuteInfo = &JobExecuteInfo{
		Job:          jobSchedulePlan.Job,
		PlanTime:     time.Now(),
		ExcutingTime: time.Now(),
	}

	return
}

func BuildJobExecuteInfo(jobSchedulePlan *JobSchedulePlan) (jobExecuteInfo *JobExecuteInfo) {

	jobExecuteInfo = &JobExecuteInfo{
		Job:          jobSchedulePlan.Job,
		PlanTime:     jobSchedulePlan.NextTime,
		ExcutingTime: time.Now(),
	}

	return
}

func (job *Job) String() string {
	return fmt.Sprintf("job:{" + "name:" + job.Name + ",command:" + job.Command + ",cronExpr:" + job.CronExpr + "}")
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
