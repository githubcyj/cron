package server

import (
	"crontab/common"
	"crontab/master"
	"crontab/master/manager"
	"crontab/model"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"mime/multipart"
	"net/http"
	"os"
)

//初始化http服务
func InitHttpApi() (err error) {
	var (
		mux *http.ServeMux
	)

	//配置路由
	mux = http.NewServeMux()
	//设置跨域
	//mux.HandleFunc("/", receiveClientRequest)
	//mux.HandleFunc("/job/save", handlerJobSave)
	mux.HandleFunc("/job/update", handlerJobUpdate)
	//mux.HandleFunc("/job/deleteForLogical", handlerJobDeleteForLogical)
	//mux.HandleFunc("/job/delete", handlerJobDelete)
	//mux.HandleFunc("/job/list", handlerJobList)
	mux.HandleFunc("/job/kill", handlerJobKill)
	mux.HandleFunc("/worker/list", handleWorkList)
	mux.HandleFunc("/upload", handUpload)

	return
}

//设置跨域
func receiveClientRequest(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Add("Access-Control-Allow-Origin", "*")
}

//更新任务
//这里接收到的参数post:job:{"name":"job","command":"echo hello","cronExpr":"* * * * *"}
func handlerJobUpdate(resp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		bytes   common.HttpReply
		postJob string
		job     common.Job
		jobArr  []byte
		newJob  *common.Job
		jobs    []*common.Job
	)

	if err = req.ParseForm(); err != nil {
		goto ERR
	}
	//获表单内容
	postJob = req.PostForm.Get("job")

	//反解析json
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		goto ERR
	}

	//更新数据库中的job
	if err = manager.GDB.UpdateJob(&job); err != nil {
		goto ERR
	}

	jobs = make([]*common.Job, 0)
	jobs = append(jobs, &job)

	//更新redis中的job
	if err = manager.GRedis.UpdateJob(jobs); err != nil {
		goto ERR
	}

	//从redis中获取最新的job
	if newJob, err = manager.GRedis.GetSingleJob(job.JobId); err != nil {
		goto ERR
	}

	//更新etcd中的job
	if _, err = manager.GJobMgr.SaveJob(newJob); err != nil {
		goto ERR
	}

	//序列化
	if jobArr, err = json.Marshal(newJob); err != nil {
		goto ERR
	}

	//返回正常应答
	bytes = common.BuildResponse(0, "success", jobArr)
	fmt.Println(bytes)

	return

ERR:
	//返回错误应答
	bytes = common.BuildResponse(-1, err.Error(), nil)
	fmt.Println(bytes)
}

//删除任务，逻辑删除，只将isDelete置为1
func HandlerJobDeleteForLogical(c *gin.Context) {
	var (
		err       error
		jobId     string
		bytes     common.HttpReply
		jobs      []*common.Job
		job       *common.Job
		deleteIds *common.DeleteIds
	)
	deleteIds = &common.DeleteIds{}
	//获得表单参数
	if err = c.ShouldBindBodyWith(deleteIds, binding.JSON); err != nil {
		goto ERR
	}

	//更新数据库
	if err = manager.GDB.DeleteJobForLogic(deleteIds.Ids); err != nil {
		goto ERR
	}

	//根据id从reids中获取对应的job
	for _, jobId = range deleteIds.Ids {
		if job, err = manager.GRedis.GetSingleJob(jobId); err != nil {
			if job == nil {
				//从数据库中获取对应的job
				if job, err = manager.GDB.GetSingleJob(jobId); err != nil {
					if job == nil {
						continue
					}
				}
			}
		}
		job.IsDel = 1
		jobs = append(jobs, job)
	}

	//更新redis
	if err = manager.GRedis.UpdateJob(jobs); err != nil {
		goto ERR
	}
	if job.Type == common.CRON_JOB_TYPE {
		//更新etcd
		if err = manager.GJobMgr.DeleteJobForLogic(deleteIds.Ids); err != nil {
			goto ERR
		}
	}
	//返回正常应答
	bytes = common.BuildResponse(0, "success", nil)
	c.JSON(http.StatusOK, gin.H{
		"data": bytes,
	})
	return

ERR:

	//返回错误应答
	bytes = common.BuildResponse(-1, err.Error(), nil)
	c.JSON(http.StatusOK, gin.H{
		"data": bytes,
	})
}

//恢复逻辑删除的任务
func HandlerJobRecover(c *gin.Context) {
	var (
		err       error
		jobId     string
		bytes     common.HttpReply
		jobs      []*common.Job
		job       *common.Job
		deleteIds *common.DeleteIds
	)
	deleteIds = &common.DeleteIds{}
	//获得表单参数
	if err = c.ShouldBindBodyWith(deleteIds, binding.JSON); err != nil {
		goto ERR
	}

	//更新数据库
	if err = manager.GDB.RecoverJob(deleteIds.Ids); err != nil {
		goto ERR
	}

	//根据id从reids中获取对应的job
	for _, jobId = range deleteIds.Ids {
		if job, err = manager.GRedis.GetSingleJob(jobId); err != nil {
			if job == nil {
				//从数据库中获取对应的job
				if job, err = manager.GDB.GetSingleJob(jobId); err != nil {
					if job == nil {
						continue
					}
				}
			}
		}
		job.IsDel = 0
		jobs = append(jobs, job)
	}

	//更新redis
	if err = manager.GRedis.UpdateJob(jobs); err != nil {
		goto ERR
	}

	//更新etcd
	if err = manager.GJobMgr.RecoverJob(deleteIds.Ids); err != nil {
		goto ERR
	}

	//返回正常应答
	bytes = common.BuildResponse(0, "success", nil)
	c.JSON(http.StatusOK, gin.H{
		"data": bytes,
	})
	return

ERR:

	//返回错误应答
	bytes = common.BuildResponse(-1, err.Error(), nil)
	c.JSON(http.StatusOK, gin.H{
		"data": bytes,
	})
}

//获取一个流水线下所有逻辑删除的job
func HandlerJobListDetele(c *gin.Context) {
	var (
		jobArr []*common.Job
		err    error
		bytes  common.HttpReply
	)

	//从redis中获得job
	if jobArr, err = manager.GRedis.GetAllDeleteJobs(); err != nil {
		goto ERR
	}

	//返回正常应答
	bytes = common.BuildResponse(0, "success", jobArr)
	c.JSON(http.StatusOK, gin.H{
		"data": bytes,
	})

	return

ERR:
	//返回错误应答
	bytes = common.BuildResponse(-1, err.Error(), nil)
	//GLogMgr.WriteLog("handlerJobSave Failed,job:" + job.Name)
	c.JSON(http.StatusOK, gin.H{
		"data": bytes,
	})
}

//强杀任务：post:name:job
func handlerJobKill(resp http.ResponseWriter, req *http.Request) {
	var (
		err        error
		jobName    string
		oldJobName common.Job
		bytes      common.HttpReply
	)
	//解析表单
	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	//获取强杀任务名
	jobName = req.PostForm.Get("name")
	if oldJobName, err = manager.GJobMgr.KillJob(jobName); err != nil {
		goto ERR
	}
	bytes = common.BuildResponse(0, "success", oldJobName)
	fmt.Println(bytes)
	return

ERR:
	//返回错误应答
	bytes = common.BuildResponse(-1, err.Error(), nil)
	fmt.Println(bytes)
	return
}

//服务列表
func handleWorkList(resp http.ResponseWriter, req *http.Request) {
	var (
		workArr []string
		err     error
		bytes   common.HttpReply
	)

	if workArr, err = manager.GWorkMgr.ListWorker(); err != nil {
		goto ERR
	}

	//返回正常应答
	bytes = common.BuildResponse(0, "success", workArr)

	fmt.Println(bytes)

ERR:
	//返回错误应答
	bytes = common.BuildResponse(-1, err.Error(), nil)
	fmt.Println(bytes)
}

func handUpload(resp http.ResponseWriter, req *http.Request) {
	var (
		err        error
		file       multipart.File
		fileHeader *multipart.FileHeader
		newFile    *os.File
		bytes      common.HttpReply
	)

	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	//从表单中读取文件
	if file, fileHeader, err = req.FormFile("file"); err != nil {
		goto ERR
	}

	//结束时关闭文件
	defer file.Close()

	//创建文件
	if newFile, err = os.Create(master.GConfig.BaseFilePath + "/" + fileHeader.Filename); err != nil {
		goto ERR
	}

	defer newFile.Close()

	//文件保存进入数据库

	bytes = common.BuildResponse(0, "success", "文件创建成功")
	fmt.Println(bytes)

ERR:
	//返回错误应答
	bytes = common.BuildResponse(-1, err.Error(), nil)
	fmt.Println(bytes)
}

//将流水线同步到etcd中，也就是开始执行任务，etcd中只要保存pipelineId即可
func HandlerSyncEtcd(c *gin.Context) {

	var (
		bytes       common.HttpReply
		err         error
		pipelineId  string
		pipelineJob *model.PipelineJob
		len         int
	)
	pipelineId = c.Query("pipelineId") //流水线号
	pipelineJob = &model.PipelineJob{
		PipelineId: pipelineId,
	}

	//判断流水线是否绑定任务
	if len, err = pipelineJob.GetRedisLen(); err != nil {
		goto ERR
	}

	//没有任务被绑定
	if len == 0 {
		//返回错误应答
		bytes = common.BuildResponse(-1, "该流水线没有绑定任务任务", nil)
		c.JSON(http.StatusBadRequest, gin.H{
			"data": bytes,
		})
	}

	//保存到etcd中
	if err = pipelineJob.SaveEtcd(); err != nil {
		goto ERR
	}

	//返回正常应答
	bytes = common.BuildResponse(0, "success", nil)
	c.JSON(http.StatusOK, gin.H{
		"data": bytes,
	})
	return

ERR:
	//返回错误应答
	bytes = common.BuildResponse(-1, err.Error(), nil)
	c.JSON(http.StatusOK, gin.H{
		"data": bytes,
	})
}
