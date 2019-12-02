package server

import (
	"crontab/common"
	"crontab/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
)

/**
  @author 胡烨
  @version 创建时间：2019/12/2 13:55
  @todo
*/

//这里接收到的参数post:job:{"name":"job","command":"echo hello","cronExpr":"* * * * *"}
func HandlerJobSave(c *gin.Context) {
	var (
		err   error
		job   model.Job
		old   *model.Job
		bytes common.HttpReply
		//postJob string
	)
	//解析post表单
	if err := c.ShouldBindBodyWith(&job, binding.JSON); err != nil {
		goto ERR
	}

	//保存job进入数据库
	if err = job.SaveDB(); err != nil {
		goto ERR
	}

	//保存job到redis中
	if err = job.SaveRedis(); err != nil {
		goto ERR
	}

	//if job.Type == common.CRON_JOB_TYPE {
	//	//将job注册到etcd中
	//	if old, err = manager.GJobMgr.SaveJob(&job); err != nil {
	//		goto ERR
	//	}
	//}
	//
	//if job.Type == common.DELAY_JOB_TYPE {
	//	if err = manager.GMgMgrProduce.PushMq(&job); err != nil {
	//		goto ERR
	//	}
	//}

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

//更新任务
func HandlerJobUpdate(c *gin.Context) {

	var (
		bytes common.HttpReply
		job   model.Job
		err   error
	)

	//解析post表单
	if err = c.ShouldBindBodyWith(&job, binding.JSON); err != nil {
		goto ERR
	}

	//更新数据库
	if err = job.UpdateJobDB(); err != nil {
		goto ERR
	}

	//更新redis
	if err = job.UpdateSingleJobRedis(); err != nil {
		goto ERR
	}

	////更新etcd
	//	//if err = manager.GJobMgr.UpdateJob(&job); err != nil {
	//	//	goto ERR
	//	//}

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
