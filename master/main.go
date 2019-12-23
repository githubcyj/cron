package main

import (
	"fmt"
	"github.com/crontab/master/config"
	_ "github.com/crontab/master/docs" //这里导入生成的docs
	"github.com/crontab/master/manager"
	"github.com/crontab/master/middleware"
	"github.com/crontab/master/router"
	"github.com/crontab/master/server"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"runtime"
)

func InitEnv() {
	//设置运行时的线程数，线程数与cpu数相等
	runtime.GOMAXPROCS(runtime.NumCPU())
}

// @title 分布式任务管理系统
// @version 1.0
// @host 127.0.0.1:8060
func main() {
	var (
		err error
	)

	//创建一个没有任何中间件的路由
	app := gin.New()
	//初始化环境
	InitEnv()

	store, _ := redis.NewStore(10, "tcp", "localhost:6381", "WEAVERemobile7*()", []byte("secret"))
	app.Use(sessions.Sessions("mysession", store))
	//加载配置
	if err = config.InitConfig(); err != nil {
		fmt.Println(err)
	}

	//启动日志管理器
	if err = manager.InitLogMgr(); err != nil {
		fmt.Println(err)
	}

	//启动http服务
	if err = server.InitHttpApi(); err != nil {
		fmt.Println(err)
	}

	//启动etcd管理器
	if err = manager.InitJobManager(); err != nil {
		fmt.Println(err)
	}

	//启动mysql
	if err = manager.InitDB(); err != nil {
		fmt.Println(err)
	}

	//启动redis
	manager.InitRedis()

	if err = manager.InitMaster(); err != nil {
		fmt.Println(err)
	}

	//启动消息队列
	//if err = manager.InitMq(); err != nil {
	//	fmt.Println(err)
	//}

	//将任务放入redis中
	//go manager.InitJobsToRedis()

	defer manager.GMasterMgr.UnVoting()

	//恢复中间件，任何panic都会被重启，并替代为500错误
	app.Use(gin.Recovery())
	//设置跨域
	app.Use(middleware.CorsMiddleWare())
	//加载路由
	router.Route(app)
	app.Run(":8060")
}
