package server

import (
	"github.com/crontab/common"
	"github.com/crontab/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
)

/**
  @author 胡烨
  @version 创建时间：2019/12/20 19:26
  @todo
*/
//增加禁止执行的任务命令
func HandlerCommandSave(c *gin.Context) {
	var (
		commandBan *model.CommandBan
		err        error
		bytes      common.HttpReply
	)
	commandBan = &model.CommandBan{}
	if err = c.ShouldBindBodyWith(commandBan, binding.JSON); err != nil {
		goto ERR
	}

	//返回正常应答
	bytes = common.BuildResponse(0, "success", commandBan)
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
