package worker

import (
	"context"
	"crontab/common"
	"math/rand"
	"os/exec"
	"time"
)

type Executor struct {
}

var (
	GExecutor *Executor
)

func InitExecute() {
	GExecutor = &Executor{}
}

//执行任务
func (executor *Executor) ExecutingJob(jobExecuteInfo *common.JobExecuteInfo) {

	var (
		cmd        *exec.Cmd
		output     []byte
		err        error
		startTime  time.Time
		endTime    time.Time
		jobResult  *common.JobExecuteResult
		cancelCtx  context.Context
		cancelFunc context.CancelFunc
	)

	cancelCtx, cancelFunc = context.WithCancel(context.TODO())

	jobExecuteInfo.CancelFunc = cancelFunc
	jobExecuteInfo.CancelCtx = cancelCtx

	//这是为了以防任务执行出错，没有输出
	jobResult = &common.JobExecuteResult{
		JobExecuteInfo: jobExecuteInfo,
		Output:         make([]byte, 0),
	}

	//创建锁
	GJobMgr.CreateLock(jobExecuteInfo.Job.Name)

	//上锁
	//随机随眠(0~1s),防止锁被一个进程占用
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

	//尝试上锁
	if err = GJobLock.TryLock(); err != nil {
		//上锁失败
		return
	}
	defer GJobLock.Unlock() //任务执行完成后释放锁

	//获得job实际运行时间
	startTime = time.Now()
	//执行任务
	cmd = exec.CommandContext(cancelCtx, "C:\\Windows\\System32\\bash.exe", "-c", jobExecuteInfo.Job.Command)
	//cmd = exec.Command("C:\\Windows\\System32\\bash.exe", "-c", jobExecuteInfo.Job.Command)

	//获得任务结束时间
	endTime = time.Now()
	//获取输出
	output, err = cmd.CombinedOutput()

	jobResult.EndTime = endTime
	jobResult.StartTime = startTime
	jobResult.Err = err
	jobResult.Output = output

	GLogMgr.WriteLog("任务执行完成，输出：" + string(output))

	//任务执行完成后，需要通知调度器任务执行完毕
	GScheduler.PushSchedulerResult(jobResult)
}
