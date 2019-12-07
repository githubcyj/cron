package worker

import (
	"crontab/common"
	"crontab/model"
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
func (executor *Executor) ExecutingJob(jobExecuteInfo *common.JobExecuteInfo, pipelineJob *model.PipelineJob) (jobRecord *model.JobRecord) {

	var (
		err error
	)
	//创建锁
	GJobMgr.CreateLock(pipelineJob.JobId)

	//上锁
	//随机随眠(0~1s),防止锁被一个进程占用
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

	//尝试上锁
	if err = GJobLock.TryLock(); err != nil {
		//上锁失败
		return
	}
	defer GJobLock.Unlock() //任务执行完成后释放锁

	if pipelineJob.Retries == 0 {
		//执行任务
		jobRecord = executor.ExecJob(jobExecuteInfo, pipelineJob.Job)
	} else {
		for i := 0; i < pipelineJob.Retries; i++ {
			//执行任务
			jobRecord = executor.ExecJob(jobExecuteInfo, pipelineJob.Job)
			if jobRecord.Status == 1 {
				break
			}
		}
	}
	return
}

func (executor *Executor) ExecJob(jobExecuteInfo *common.JobExecuteInfo, Job *model.Job) (jobRecord *model.JobRecord) {
	var (
		cmd       *exec.Cmd
		output    []byte
		err       error
		startTime time.Time
		endTime   time.Time
		resChan   chan struct {
			output string
			err    error
		}
		res struct {
			output string
			err    error
		}
	)
	jobRecord = &model.JobRecord{}
	startTime = time.Now()
	resChan = make(chan struct {
		output string
		err    error
	})
	go func() {
		cmd = exec.CommandContext(jobExecuteInfo.CancelCtx, "C:\\Windows\\System32\\bash.exe", "-c", Job.Command)
		output, err = cmd.CombinedOutput()
		resChan <- struct {
			output string
			err    error
		}{output: string(output), err: err}
	}()
	endTime = time.Now()
	res = <-resChan
	GLogMgr.WriteLog("任务执行完成，输出：" + res.output)
	jobRecord.Content = Job.Command
	jobRecord.Result = res.output
	jobRecord.JobId = Job.JobId
	jobRecord.BeginWith = startTime
	jobRecord.FinishWith = endTime
	jobRecord.Duration = int(endTime.Sub(startTime).Seconds())
	if res.err != nil {
		jobRecord.Status = 0
	} else {
		jobRecord.Status = 1
	}
	return
}
