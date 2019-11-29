package router

import (
	"crontab/master/server"
	"github.com/gin-gonic/gin"
)

//配置路由
func Route(router *gin.Engine) {

	api := router.Group("/job")
	{
		api.POST("/list", server.HandlerJobList)                         //获取所有未删除的job
		api.POST("/list/delete", server.HandlerJobListDetele)            //获取所有逻辑删除的job
		api.POST("/save", server.HandlerJobSave)                         //保存job
		api.POST("/deleteForLogical", server.HandlerJobDeleteForLogical) //逻辑删除job
		api.POST("/delete", server.HandlerJobDelete)                     //物理删除job
		api.POST("/recoverJob", server.HandlerJobRecover)                //恢复逻辑删除的job
		api.POST("/update", server.HandlerJobUpdate)                     //更新任务

		api.POST("/create/pipelines", server.HandlerPipeCreate) //创建流水线
		api.POST("/pipeline/job", server.HandlerPipeCreate)     //任务与流水线关联
	}
}
