package log

import (
	"crypto/tls"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/weekface/mgorus"
	"gopkg.in/mgo.v2"
	"net"
	"time"
)

type LogMgr struct {
	Log *logrus.Logger
}

var (
	GLogMgr *LogMgr
)

func InitLogMgr() (err error) {

	var (
		log    *logrus.Logger
		logMgr *LogMgr
		//fileInfo os.FileInfo
	)

	s, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{"localhost:27017"},
		Timeout:  5 * time.Second,
		Database: "admin",
		Username: "",
		Password: "",
		DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), &tls.Config{InsecureSkipVerify: true})
			return conn, err
		},
	})
	if err != nil {
		log.Fatalf("can't create session: %s\n", err)
	}

	c := s.DB("cron").C("log")
	hooker := mgorus.NewHookerFromCollection(c)
	if err == nil {
		log.Hooks.Add(hooker)
	} else {
		fmt.Print(err)
	}

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

func (logMgr *LogMgr) WriteLog() {

}
