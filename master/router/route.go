package router

import (
	"crontab/master/server"
	"github.com/gin-gonic/gin"
)

//配置路由
func Route(router *gin.Engine) {

	api := router.Group("/cron")
	{
		//api.POST("/record/delete", server.HandlerJobListDetele) //获取一个流水线下所有逻辑删除的job
		//
		//api.POST("/recoverJob", server.HandlerJobRecover) //恢复逻辑删除的job
		//api.POST("/job/update", server.HandlerJobUpdate) //更新任务
		//api.POST("/record/deleteForLogical", server.HandlerJobDeleteForLogical) //逻辑删除job

		api.GET("/record/list", server.HandlerJobList)   //获取一个流水线下所有的job
		api.POST("/job/delete", server.HandlerJobDelete) //物理删除一个流水线下的job
		api.POST("/job/save", server.HandlerJobSave)     //创建job

		api.POST("/create/pipelines", server.HandlerPipeCreate) //创建流水线
		api.POST("/pipeline/job", server.HandlerPiplineJob)     //任务与流水线关联
		api.GET("/pipeline/steps", server.HandlerStep)          //任务排序

		//api.GET("/pipeline/run", server.HandlerSyncEtcd)      //将流水线同步到etcd中开始调度
		api.GET("/pipeline/kill", server.HandlerPipelineKill) //将流水线同步到etcd中开始调度
	}
}
