package server

import (
	"crontab/common"
	"crontab/entity"
	"crontab/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
)

/**
  @author 胡烨
  @version 创建时间：2019/12/2 14:09
  @todo
*/

//获取一个流水线下所有未删除的job
func HandlerJobList(c *gin.Context) {
	var (
		jobArr      []*model.Job
		err         error
		bytes       common.HttpReply
		pipelineId  string
		pipelineJob *model.PipelineJob
	)

	pipelineId = c.Query("pipelineId")
	pipelineJob = &model.PipelineJob{PipelineId: pipelineId}

	//从redis中获得job
	if jobArr, err = pipelineJob.GetAllJobRedis(); err != nil {
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

//物理删除一个流水线下的job
func HandlerJobDelete(c *gin.Context) {
	var (
		err error
		//jobId string
		deleteIds *entity.DeleteIds
		bytes     common.HttpReply
		ids       []string
	)
	deleteIds = &entity.DeleteIds{}
	//获得表单参数
	if err = c.ShouldBindBodyWith(deleteIds, binding.JSON); err != nil {
		goto ERR
	}
	//ids = strings.Split(jobId, ",")

	//从redis中删除
	if err = deleteIds.DelJobsRedis(); err != nil {
		goto ERR
	}

	//从数据库中删除
	if err = deleteIds.DelJobsDB(); err != nil {
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

//强杀流水线
func HandlerPipelineKill(c *gin.Context) {
	var (
		pipelineId  string
		pipelineJob *model.PipelineJob
		err         error
		bytes       common.HttpReply
	)

	pipelineId = c.Query("pipelineId")
	pipelineJob = &model.PipelineJob{PipelineId: pipelineId}
	if err = pipelineJob.KillEtcd(); err != nil {
		goto ERR
	}

	//返回正常应答
	bytes = common.BuildResponse(0, "success", pipelineJob)
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
