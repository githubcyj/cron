package server

import (
	"crontab/common"
	"crontab/constants"
	"crontab/master/manager"
	"crontab/model"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"net/http"
)

/**
  @author 胡烨
  @version 创建时间：2019/12/9 16:08
  @todo
*/

func HandlerJobFile(c *gin.Context) {
	var (
		file      *multipart.FileHeader
		err       error
		path      string
		bytes     common.HttpReply
		fileModel *model.File
		fileId    string
	)
	if file, err = c.FormFile("file"); err != nil {
		manager.GLogMgr.WriteLog("文件上传失败：" + err.Error())
		goto ERR
	}

	fileModel = &model.File{Name: file.Filename}

	//插入数据库
	if fileId, err = fileModel.SaveDb(); err != nil {
		goto ERR
	}
	path = constants.FILE_PATH + fileId
	if err = c.SaveUploadedFile(file, path); err != nil {
		manager.GLogMgr.WriteLog(err.Error())
		goto ERR
	}
	//插入redis
	fileModel.FileId = fileId
	if err = fileModel.SaveRedis(); err != nil {
		goto ERR
	}
	//返回正常应答
	bytes = common.BuildResponse(0, "success", fileId)
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
