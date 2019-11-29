package master

import (
	"crontab/common"
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
	if err = GDB.UpdateJob(&job); err != nil {
		goto ERR
	}

	jobs = make([]*common.Job, 0)
	jobs = append(jobs, &job)

	//更新redis中的job
	if err = GRedis.UpdateJob(jobs); err != nil {
		goto ERR
	}

	//从redis中获取最新的job
	if newJob, err = GRedis.GetSingleJob(job.JobId); err != nil {
		goto ERR
	}

	//更新etcd中的job
	if _, err = GJobMgr.SaveJob(newJob); err != nil {
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

//这里接收到的参数post:job:{"name":"job","command":"echo hello","cronExpr":"* * * * *"}
func HandlerJobSave(c *gin.Context) {
	var (
		err   error
		job   common.Job
		old   *common.Job
		bytes common.HttpReply
		//postJob string
	)
	//解析post表单
	if err := c.ShouldBindBodyWith(&job, binding.JSON); err != nil {
		goto ERR
	}

	//保存job进入数据库
	if err = GDB.SaveJob(&job); err != nil {
		goto ERR
	}

	//保存job到redis中
	if err = GRedis.AddJob(&job); err != nil {
		goto ERR
	}
	if job.Type == common.CRON_JOB_TYPE {
		//将job注册到etcd中
		if old, err = GJobMgr.SaveJob(&job); err != nil {
			goto ERR
		}
	}

	if job.Type == common.DELAY_JOB_TYPE {
		if err = GMgMgrProduce.PushMq(&job); err != nil {
			goto ERR
		}
	}

	//返回正常应答
	bytes = common.BuildResponse(0, "success", old)
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
	if err = GDB.DeleteJobForLogic(deleteIds.Ids); err != nil {
		goto ERR
	}

	//根据id从reids中获取对应的job
	for _, jobId = range deleteIds.Ids {
		if job, err = GRedis.GetSingleJob(jobId); err != nil {
			if job == nil {
				//从数据库中获取对应的job
				if job, err = GDB.GetSingleJob(jobId); err != nil {
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
	if err = GRedis.UpdateJob(jobs); err != nil {
		goto ERR
	}
	if job.Type == common.CRON_JOB_TYPE {
		//更新etcd
		if err = GJobMgr.DeleteJobForLogic(deleteIds.Ids); err != nil {
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

//更新任务
func HandlerJobUpdate(c *gin.Context) {

	var (
		bytes common.HttpReply
		job   common.Job
		err   error
	)

	//解析post表单
	if err = c.ShouldBindBodyWith(&job, binding.JSON); err != nil {
		goto ERR
	}

	//更新数据库
	if err = GDB.UpdateJob(&job); err != nil {
		goto ERR
	}

	//更新redis
	if err = GRedis.UpdateSingleJob(&job); err != nil {
		goto ERR
	}

	//更新etcd
	if err = GJobMgr.UpdateJob(&job); err != nil {
		goto ERR
	}

	//返回正常应答
	bytes = common.BuildResponse(0, "success", nil)
	c.JSON(http.StatusOK, gin.H{
		"data": bytes,
	})

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
	if err = GDB.RecoverJob(deleteIds.Ids); err != nil {
		goto ERR
	}

	//根据id从reids中获取对应的job
	for _, jobId = range deleteIds.Ids {
		if job, err = GRedis.GetSingleJob(jobId); err != nil {
			if job == nil {
				//从数据库中获取对应的job
				if job, err = GDB.GetSingleJob(jobId); err != nil {
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
	if err = GRedis.UpdateJob(jobs); err != nil {
		goto ERR
	}

	//更新etcd
	if err = GJobMgr.RecoverJob(deleteIds.Ids); err != nil {
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

//删除任务，post:jobId:["fasdafd"],物理删除
func HandlerJobDelete(c *gin.Context) {
	var (
		err error
		//jobId string
		deleteIds *common.DeleteIds
		bytes     common.HttpReply
		ids       []string
	)
	deleteIds = &common.DeleteIds{}
	//获得表单参数
	if err = c.ShouldBindBodyWith(deleteIds, binding.JSON); err != nil {
		goto ERR
	}
	//ids = strings.Split(jobId, ",")

	//从redis中删除
	if err = GRedis.DelJobs(deleteIds.Ids); err != nil {
		goto ERR
	}

	//从数据库中删除
	if err = GDB.DelJob(deleteIds.Ids); err != nil {
		goto ERR
	}
	//从etcd中删除
	if ids, err = GJobMgr.DelJob(deleteIds.Ids); err != nil {
		goto ERR
	}

	//返回正常应答
	bytes = common.BuildResponse(0, "success", ids)
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

//获取所有的job
func HandlerJobList(c *gin.Context) {
	var (
		jobArr []*common.Job
		err    error
		bytes  common.HttpReply
	)

	//从redis中获得job
	if jobArr, err = GRedis.GetAllJobs(); err != nil {
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

//获取所有逻辑删除的job
func HandlerJobListDetele(c *gin.Context) {
	var (
		jobArr []*common.Job
		err    error
		bytes  common.HttpReply
	)

	//从redis中获得job
	if jobArr, err = GRedis.GetAllDeleteJobs(); err != nil {
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
	if oldJobName, err = GJobMgr.KillJob(jobName); err != nil {
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

	if workArr, err = GWorkMgr.ListWorker(); err != nil {
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
	if newFile, err = os.Create(GConfig.BaseFilePath + "/" + fileHeader.Filename); err != nil {
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
