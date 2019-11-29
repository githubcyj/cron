package router

import (
	"crontab/master"
	"github.com/gin-gonic/gin"
)

//配置路由
func Route(router *gin.Engine) {

	api := router.Group("/job")
	{
		api.POST("/list", master.HandlerJobList)                         //获取所有未删除的job
		api.POST("/list/delete", master.HandlerJobListDetele)            //获取所有逻辑删除的job
		api.POST("/save", master.HandlerJobSave)                         //保存job
		api.POST("/deleteForLogical", master.HandlerJobDeleteForLogical) //逻辑删除job
		api.POST("/delete", master.HandlerJobDelete)                     //物理删除job
		api.POST("/recoverJob", master.HandlerJobRecover)                //恢复逻辑删除的job
		api.POST("/update", master.HandlerJobUpdate)                     //更新任务
	}
}
