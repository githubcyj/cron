package main

import (
	"crontab/master"
	"crontab/master/middleware"
	"crontab/master/router"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"runtime"
)

func InitEnv() {
	//设置运行时的线程数，线程数与cpu数相等
	runtime.GOMAXPROCS(runtime.NumCPU())
}

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
	if err = master.InitConfig(); err != nil {
		fmt.Println(err)
	}

	//启动日志管理器
	if err = master.InitLogMgr(); err != nil {
		fmt.Println(err)
	}

	//启动http服务
	if err = master.InitHttpApi(); err != nil {
		fmt.Println(err)
	}

	//启动etcd管理器
	if err = master.InitJobManager(); err != nil {
		fmt.Println(err)
	}

	//启动mysql
	if err = master.InitDB(); err != nil {
		fmt.Println(err)
	}

	//启动redis
	master.InitRedis()

	if err = master.InitMaster(); err != nil {
		fmt.Println(err)
	}

	//启动消息队列
	if err = master.InitMq(); err != nil {
		fmt.Println(err)
	}

	//将任务放入redis中
	go master.InitJobsToRedis()

	defer master.GMasterMgr.UnVoting()

	//恢复中间件，任何panic都会被重启，并替代为500错误
	app.Use(gin.Recovery())
	//设置跨域
	app.Use(middleware.CorsMiddleWare())
	//加载路由
	router.Route(app)
	app.Run(":8060")
}
