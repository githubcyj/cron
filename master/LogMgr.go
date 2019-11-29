package master

import (
	"github.com/sirupsen/logrus"
	"os"
)

type LogMgr struct {
	Log *logrus.Logger
}

var (
	GLogMgr *LogMgr
)

func InitLogMgr() (err error) {

	var (
		file   *os.File
		log    *logrus.Logger
		logMgr *LogMgr
		//fileInfo os.FileInfo
	)

	//获取文件信息
	_, err = os.Stat(GConfig.LogFilepath + GConfig.LogFilename)

	//判断文件是否存在
	if os.IsNotExist(err) { //文件不存在
		//创建文件
		if file, err = os.Create(GConfig.LogFilepath + GConfig.LogFilename); err != nil { //文件创建失败
			return
		}
	}
	defer file.Close()

	log = logrus.New()
	//设置日志输出格式
	log.SetFormatter(&logrus.TextFormatter{})
	//设置日志级别
	log.Level = logrus.DebugLevel

	logMgr = &LogMgr{Log: log}

	//赋值单例
	GLogMgr = logMgr

	return
}

//向文件中写日志
func (logMgr *LogMgr) WriteLog(msg string) {
	//var (
	//	file *os.File
	//	err  error
	//)
	//打开文件
	//if file, err = os.OpenFile(GConfig.LogFilepath+GConfig.LogFilename, os.O_CREATE|os.O_WRONLY, 0666); err == nil {
	//	logMgr.Log.Out = file
	//} else {
	//	logMgr.Log.Info("写入日志到文件失败")
	//}
	logMgr.Log.Info(msg)
	return
}
