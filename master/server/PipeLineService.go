package server

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/crontab/common"
	"github.com/crontab/constants"
	"github.com/crontab/master/manager"
	"github.com/crontab/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
	"sort"
	"strconv"
)

/**
  @author 胡烨
  @version 创建时间：2019/12/2 13:53
  @todo
*/

//创建流水线
func HandlerPipeCreate(c *gin.Context) {
	var (
		pipeline  model.Pipeline
		bytes     common.HttpReply
		err       error
		job       *model.Job
		jobFinish *model.Job
		jobFail   *model.Job
	)

	//解析post表单
	if err := c.ShouldBindBodyWith(&pipeline, binding.JSON); err != nil {
		goto ERR
	}

	//根据finished failed获取job
	//从redis中获取数据
	if pipeline.Finished != "" {
		job = &model.Job{JobId: pipeline.Finished}
		if jobFinish, err = job.GetSingleJobRedis(); err != nil {
			if jobFinish == nil {
				if jobFinish, err = job.GetSingleJobDB(); err != nil {
					goto ERR
				}
			}
		}
	}
	if pipeline.Failed != "" {
		job = &model.Job{JobId: pipeline.Failed}
		if jobFail, err = job.GetSingleJobRedis(); err != nil {
			if jobFail == nil {
				//从数据库获取
				if jobFail, err = job.GetSingleJobDB(); err != nil {
					goto ERR
				}
			}
		}
	}

	pipeline.FinishedJob = jobFinish
	pipeline.FailedJob = jobFail

	//加入数据库
	if err = pipeline.SaveDB(); err != nil {
		goto ERR
	}

	//加入redis
	if err = pipeline.SaveRedis(); err != nil {
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
	c.JSON(http.StatusOK, gin.H{
		"data": bytes,
	})
}

//任务与流水线关联
func HandlerPiplineJob(c *gin.Context) {
	var (
		bytes          common.HttpReply
		pipelineJob    model.PipelineJob
		pipelineJobArr []*model.PipelineJob
		err            error
		len            int
		job            *model.Job
		jobR           *model.Job
		pipeline       *model.Pipeline
		pipelineR      *model.Pipeline
	)

	//解析post表单
	if err = c.ShouldBindBodyWith(&pipelineJob, binding.JSON); err != nil {
		goto ERR
	}

	if len, err = pipelineJob.GetPipelineJobLen(); err != nil {
		goto ERR
	}

	pipelineJob.Step = len + 1

	//根据jobid获取job
	job = &model.Job{JobId: pipelineJob.JobId}
	if jobR, err = job.GetSingleJobRedis(); err != nil {
		goto ERR
	}
	if jobR == nil {
		if jobR, err = job.GetSingleJobDB(); err != nil {
			goto ERR
		}
		if jobR == nil {
			err = errors.New("任务不存在")
			goto ERR
		}
	}
	//保存数据库
	if err = pipelineJob.SaveDB(); err != nil {
		goto ERR
	}

	pipelineJob.Job = jobR
	//保存redis
	if err = pipelineJob.SaveRedis(); err != nil {
		goto ERR
	}

	//更新redis中pipeline中的信息
	//1. 从redis中获取
	pipeline = &model.Pipeline{PipelineId: pipelineJob.PipelineId}
	if pipelineR, err = pipeline.GetRedis(); err != nil {
		goto ERR
	}
	//2. 修改Steps内容
	if pipelineJobArr, err = pipelineJob.GetAllJobRedis(); err != nil {
		goto ERR
	}
	pipelineR.Steps = pipelineJobArr

	//3. 删除原本redis中的数据
	if err = pipelineR.DelRedis(); err != nil {
		goto ERR
	}

	//重新存入redis
	if err = pipelineR.SaveRedis(); err != nil {
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
	c.JSON(http.StatusOK, gin.H{
		"data": bytes,
	})
}

//任务排序
func HandlerStep(c *gin.Context) {
	var (
		bytes       common.HttpReply
		relations   []*model.PipelineJob
		rows        *sql.Rows
		err         error
		job_id      string
		step        int
		pipelineId  string
		current     int
		origin      int
		pipelineJob *model.PipelineJob
		count       int
		job         model.Job
	)
	pipelineId = c.Query("pipelineId")            //流水线号
	current, _ = strconv.Atoi(c.Query("current")) //挪到之后的位置
	origin, _ = strconv.Atoi(c.Query("origin"))   //挪动之前的位置
	relations = make([]*model.PipelineJob, 0)

	if rows, err = manager.GDB.DB.Table("pipeline_jobs").Select("pipeline_jobs.job_id,pipeline_jobs.step,pipeline_jobs.pipeline_id").
		Joins("inner join jobs on jobs.job_id = pipeline_jobs.job_id and pipeline_jobs.pipeline_id=?", pipelineId).
		Order("pipeline_jobs.step").Rows(); err != nil {
		goto ERR
	}

	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&job_id, &step, &pipelineId); err != nil {
			fmt.Println(err.Error())
			goto ERR
		}
		pipelineJob = &model.PipelineJob{
			PipelineId: pipelineId,
			JobId:      job_id,
			Step:       step,
		}
		relations = append(relations, pipelineJob)
	}
	count = len(relations)

	//删除redis中的数据
	if err = pipelineJob.DelRedis(); err != nil {
		goto ERR
	}

	//从任意位置挪到第一个位置
	if current == 0 && origin > 0 {
		for i := 0; i < count; i++ {
			if i < origin {
				relations[i].Step++
				if err = manager.GDB.DB.Model(&model.PipelineJob{}).
					Where("job_id=?", relations[i].JobId).
					Update("step", relations[i].Step).Error; err != nil {
					manager.GLogMgr.WriteLog("更新排序错误：" + err.Error())
					goto ERR
				}
			}
		}
	}

	//从上往下挪动
	if origin < current {
		for i := 0; i < count; i++ {
			if i > origin && i <= current {
				relations[i].Step--
				if err = manager.GDB.DB.Model(&model.PipelineJob{}).
					Where("job_id=?", relations[i].JobId).
					Update("step", relations[i].Step).Error; err != nil {
					manager.GLogMgr.WriteLog("更新排序错误：" + err.Error())
					goto ERR
				}
			}
		}
	}

	//从下往上挪动
	if current < origin {
		for i := 0; i < count; i++ {
			if i >= current && i < origin {
				relations[i].Step++
				if err = manager.GDB.DB.Model(&model.PipelineJob{}).
					Where("job_id=?", relations[i].JobId).
					Update("step", relations[i].Step).Error; err != nil {
					manager.GLogMgr.WriteLog("更新排序错误：" + err.Error())
					goto ERR
				}
			}
		}
	}

	//更新移动的任务
	if err = manager.GDB.DB.Model(&model.PipelineJob{}).
		Where("job_id=?", relations[origin].JobId).
		Update("step", current).Error; err != nil {
		manager.GLogMgr.WriteLog("更新排序错误：" + err.Error())
		goto ERR
	}

	//排序
	sort.Slice(relations, func(before, after int) bool {
		return relations[before].Step < relations[after].Step
	})

	//获取任务详情
	for i := 0; i < count; i++ {
		if err = manager.GDB.DB.Where("job_id = ?", relations[i].JobId).First(&job).Error; err != nil {
			manager.GLogMgr.WriteLog("获取单个任务失败：" + err.Error())
			goto ERR
		}
		relations[i].Job = &job

		//保存redis
		if err = relations[i].SaveRedis(); err != nil {
			goto ERR
		}
	}

	//返回正常应答
	bytes = common.BuildResponse(0, "success", relations)
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

//将流水线同步到etcd中开始调度
func HandlerSyncEtcd(c *gin.Context) {
	var (
		pipelineId string
		pipeline   *model.Pipeline
		pipelineR  *model.Pipeline
		bytes      common.HttpReply
		err        error
	)

	pipelineId = c.Query("pipelineId")
	pipeline = &model.Pipeline{PipelineId: pipelineId}
	//从redis中获取对应的流水线
	if pipelineR, err = pipeline.GetRedis(); err != nil {
		goto ERR
	}
	//定时任务加入etcd立即执行
	if pipelineR.Type == constants.CRON_JOB_TYPE {
		if err = pipelineR.SaveEtcd(); err != nil {
			goto ERR
		}
	}

	//延时任务，则加入mq中等待延时时间到达执行
	if pipelineR.Type == constants.DELAY_JOB_TYPE {
		if err = pipelineR.PushMq(); err != nil {
			goto ERR
		}
	}

	//返回正常应答
	bytes = common.BuildResponse(0, "success", pipelineR)
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

//删除流水线
func HandlerPipeDelete(c *gin.Context) {
	var (
		pipelineId  string
		pipeline    *model.Pipeline
		pipelineJob *model.PipelineJob
		err         error
		bytes       common.HttpReply
	)

	pipelineId = c.Query("pipelineId")
	pipeline = &model.Pipeline{PipelineId: pipelineId}
	pipelineJob = &model.PipelineJob{PipelineId: pipelineId}

	//先删除redis中流水线
	if err = pipeline.DelRedis(); err != nil {
		goto ERR
	}

	//删除数据库中流水线
	if err = pipeline.DelDB(); err != nil {
		goto ERR
	}

	//删除流水线与任务关联
	if err = pipelineJob.DelRedis(); err != nil {
		goto ERR
	}

	if err = pipelineJob.DelDB(); err != nil {
		goto ERR
	}

	//返回正常应答
	bytes = common.BuildResponse(0, "success", pipelineId)
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
