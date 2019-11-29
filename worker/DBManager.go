package worker

import (
	"crontab/common"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

/**
  @author 胡烨
  @version 创建时间：2019/11/18 11:27
  @todo 数据库连接
*/
/**
数据库连接相关
*/

type DBManager struct {
	DB *gorm.DB
}

var GDB *DBManager

//初始化数据库连接
func InitDB() (err error) {
	var (
		url string
		db  *gorm.DB
	)
	url = fmt.Sprintf("%s:%s@(%s:%d)/%s?allowNativePasswords=true&parseTime=True&loc=Local",
		GConfig.User, GConfig.Password, GConfig.Host, GConfig.Port, GConfig.Database)
	if db, err = gorm.Open(GConfig.Dialect, url); err != nil {
		return
	}
	db.DB().SetMaxIdleConns(GConfig.MaxIdleConns)
	db.DB().SetMaxOpenConns(GConfig.MaxOpenConns)
	GDB = &DBManager{DB: db}
	return
}

//任务保存进入数据库
func (db *DBManager) SaveJob(job *common.Job) (err error) {
	if err = db.DB.Create(job).Error; err != nil {
		GLogMgr.WriteLog("插入数据失败，失败原因：" + err.Error())
		return
	}
	return
}

func (db *DBManager) GetSingleJob(jobId string) (job *common.Job, err error) {
	job = &common.Job{}
	if err = db.DB.Where("jobId = ?", jobId).First(&job).Error; err != nil {
		return job, err
	}

	return job, err
}
