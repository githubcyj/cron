package router

import (
	_ "github.com/crontab/master/docs" //这里导入生成的docs
	"github.com/crontab/master/server"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

//配置路由
func Route(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	api := router.Group("/cron")
	{
		//api.POST("/record/delete", server.HandlerJobListDetele) //获取一个流水线下所有逻辑删除的job
		//
		//api.POST("/recoverJob", server.HandlerJobRecover) //恢复逻辑删除的job
		//api.POST("/job/update", server.HandlerJobUpdate) //更新任务
		//api.POST("/record/deleteForLogical", server.HandlerJobDeleteForLogical) //逻辑删除job

		api.GET("/record/list", server.HandlerJobList)        //获取一个流水线下所有的job
		api.POST("/job/delete", server.HandlerJobDelete)      //物理删除一个流水线下的job
		api.POST("/job/save", server.HandlerJobSave)          //创建job
		api.POST("/job/save/file", server.HandlerJobSaveFile) //创建job,job内容为文件
		api.POST("/job/uploader", server.HandlerJobFile)      //上传文件

		api.POST("/create/pipelines", server.HandlerPipeCreate) //创建流水线
		api.GET("/delete/pipelines", server.HandlerPipeDelete)  //删除流水线
		api.POST("/pipeline/job", server.HandlerPiplineJob)     //任务与流水线关联
		api.GET("/pipeline/steps", server.HandlerStep)          //任务排序

		api.GET("/pipeline/run", server.HandlerSyncEtcd)      //将流水线同步到etcd中开始调度
		api.GET("/pipeline/kill", server.HandlerPipelineKill) //将流水线同步到etcd中开始调度

		api.GET("/node/list", server.HandlerNodeList)               //获得所有的node节点，包括在线与不在线
		api.POST("/node/pipeline/bind", server.HandlerNodePipeline) //流水线与node节点绑定

		api.POST("/command/ban/add", server.HandlerCommandSave) //增加禁止执行的任务命令
	}
}
