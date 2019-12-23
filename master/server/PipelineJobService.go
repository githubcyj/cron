package server

import (
	"errors"
	"github.com/crontab/common"
	"github.com/crontab/entity"
	"github.com/crontab/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
)

/**
  @author 胡烨
  @version 创建时间：2019/12/2 14:09
  @todo
*/

// @Summary 获取一个流水线下所有未删除的job
// @Produce json
// @Param studName query string true "学生姓名"
// @Success 200 {object} HttpReply
// @Router /record/list [post]
func HandlerJobList(c *gin.Context) {
	var (
		jobArr         []*model.Job
		pipelineJobArr []*model.PipelineJob
		pipelineJobR   *model.PipelineJob
		err            error
		bytes          common.HttpReply
		pipelineId     string
		pipelineJob    *model.PipelineJob
		job            *model.Job
		jobR           *model.Job
	)

	pipelineId = c.Query("pipelineId")
	pipelineJob = &model.PipelineJob{PipelineId: pipelineId}

	//从redis中获得job
	if pipelineJobArr, err = pipelineJob.GetAllJobRedis(); err != nil {
		goto ERR
	}

	for _, pipelineJobR = range pipelineJobArr {
		job = &model.Job{JobId: pipelineJobR.JobId}
		if jobR, err = job.GetSingleJobRedis(); err != nil {
			goto ERR
		}
		if jobR == nil {
			if jobR, err = job.GetSingleJobDB(); err != nil {
				goto ERR
			}
		}
		jobArr = append(jobArr, jobR)
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
		deleteIds   *entity.DeleteIds
		bytes       common.HttpReply
		ids         []string
		id          string
		pipelineJob *model.PipelineJob
		bind        bool
	)
	deleteIds = &entity.DeleteIds{}
	//获得表单参数
	if err = c.ShouldBindBodyWith(deleteIds, binding.JSON); err != nil {
		goto ERR
	}
	//ids = strings.Split(jobId, ",")
	ids = make([]string, 0)
	//判断任务是否和流水线进行绑定，如果绑定则先解绑才能删除
	for _, id = range deleteIds.JobIds {
		pipelineJob = &model.PipelineJob{JobId: id}
		if bind, err = pipelineJob.IsJobBindPipeline(); err != nil {
			goto ERR
		}
		if bind == false {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		err = errors.New("任务都已绑定流水线，无法删除")
		goto ERR
	}

	deleteIds.JobIds = ids
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
		pipelineId string
		pipeline   *model.Pipeline
		err        error
		bytes      common.HttpReply
	)

	pipelineId = c.Query("pipelineId")
	pipeline = &model.Pipeline{PipelineId: pipelineId}
	if err = pipeline.KillEtcd(); err != nil {
		goto ERR
	}

	//返回正常应答
	bytes = common.BuildResponse(0, "success", pipeline)
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
