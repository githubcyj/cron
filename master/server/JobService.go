package server

import (
	"errors"
	"fmt"
	"github.com/crontab/common"
	"github.com/crontab/constants"
	"github.com/crontab/model"
	"github.com/crontab/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
	"strings"
)

/**
  @author 胡烨
  @version 创建时间：2019/12/2 13:55
  @todo
*/

//这里接收到的参数post:job:{"name":"job","command":"echo hello","cronExpr":"* * * * *"}
func HandlerJobSave(c *gin.Context) {
	var (
		err         error
		job         model.Job
		old         *model.Job
		bytes       common.HttpReply
		file        *model.File
		isExist     bool
		commandList []*model.CommandBan
		command     *model.CommandBan
		file        *model.File
		fileR       *model.File
		banExist    bool
		//postJob string
	)
	//解析post表单
	if err := c.ShouldBindBodyWith(&job, binding.JSON); err != nil {
		goto ERR
	}
	if job.IsFile == 1 {
		file = &model.File{FileId: job.FileId}
		//需要判断文件是否存在
		if isExist = file.IsExist(); !isExist {
			err = errors.New("文件不存在，请重新上传")
			goto ERR
		}
	}
	command = &model.CommandBan{}
	//获取所有禁止任务
	if commandList, err = command.GetCommand(); err != nil {
		goto ERR
	}
	//判断任务中是否有禁止命令
	if job.IsFile == 1 { //任务是文件任务，则需要判断文件中是否包含禁止命令
		//获取文件名
		file = &model.File{FileId: job.FileId}
		if fileR, err = file.GetSingleRedis(); err != nil {
			goto ERR
		}
		for _, command = range commandList {
			if util.IsFindStrInFile(command.Command, constants.FILE_PATH+fileR.Name) {
				fmt.Println("文件中包含禁止命令，此任务将不会执行")
				job.Status = -1
			}

		}
	} else {
		for _, command = range commandList {
			if strings.Contains(job.Command, command.Command) {
				fmt.Println("文件中包含禁止命令，此任务将不会执行")
				job.Status = -1
			}
		}
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
func HandlerJobSaveFile(c *gin.Context) {
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
