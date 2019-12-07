package main

import (
	"crontab/worker"
	"fmt"
	"runtime"
	"time"
)

func InitEnv() {
	//设置运行时的线程数，线程数与cpu数相等
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var (
		err error
	)

	InitEnv()
	//加载配置
	if err = worker.InitConfig(); err != nil {
		goto ERR
	}

	//启动redis连接
	worker.InitRedis()

	//启动数据库连接
	if err = worker.InitDB(); err != nil {
		goto ERR
	}

	//启动日志管理器
	if err = worker.InitLogMgr(); err != nil {
		goto ERR
	}

	//启动调度器
	if err = worker.InitScheduler(); err != nil {
		goto ERR
	}

	//启动任务管理器
	if err = worker.InitJobMgr(); err != nil {
		goto ERR
	}

	//启动任务调度器
	worker.InitExecute()

	//启动日志
	if err = worker.InitLogMgr(); err != nil {
		goto ERR
	}

	//启动延时任务监听
	//if err = worker.InitMq(); err != nil {
	//	goto ERR
	//}

	for {
		time.Sleep(1 * time.Second)
	}

	//启动服务注册
	if err = worker.InitRegister(); err != nil {
		goto ERR
	}

ERR:
	fmt.Println(err)
}
