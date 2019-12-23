package manager

import (
	"fmt"
	"github.com/crontab/master/config"
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
		config.GConfig.User, config.GConfig.Password, config.GConfig.Host, config.GConfig.Port, config.GConfig.Database)
	if db, err = gorm.Open(config.GConfig.Dialect, url); err != nil {
		return
	}
	db.DB().SetMaxIdleConns(config.GConfig.MaxIdleConns)
	db.DB().SetMaxOpenConns(config.GConfig.MaxOpenConns)
	GDB = &DBManager{DB: db}
	return
}

//
////更新job
//func (db *DBManager) UpdateJob(job *common.Job) (err error) {
//	var (
//		oldJob *common.Job
//	)
//	//从redis中查出该job
//	if oldJob, err = GRedis.GetSingleJob(job.JobId); err != nil {
//		//redis中没有对应的job
//		if oldJob == nil {
//			//从数据库查询对应的job
//			oldJob, _ = db.GetSingleJob(job.JobId)
//			//数据库中没有对应的内容，则需要插入数据
//			if oldJob == nil {
//				if err = db.saveJob(job); err != nil {
//					return
//				} else {
//					return nil
//				}
//			}
//		} else {
//			return
//		}
//	}
//
//	//获取旧job updateCount并加1
//	job.UpdateCount = oldJob.UpdateCount + 1
//	job.Id = oldJob.Id
//
//	err = db.DB.Save(job).Error
//	return err
//}
//
//func (db *DBManager) DelJob(jobIds []string) (err error) {
//	var (
//		jobId string
//	)
//	for _, jobId = range jobIds {
//		if err = db.DB.Where("job_id = ?", jobId).Delete(common.Job{}).Error; err != nil {
//			return
//		}
//	}
//
//	return
//}
//
////从数据库中查出所有任务
//func (db *DBManager) ListJob() (jobs []*common.Job, err error) {
//
//	jobs = make([]*common.Job, 0)
//
//	if err = db.DB.Find(&jobs).Error; err != nil {
//		return jobs, err
//	}
//	return jobs, err
//}
//
//func (db *DBManager) DeleteJobForLogic(jobIds []string) (err error) {
//	var (
//		jobId string
//		job   *common.Job
//	)
//
//	for _, jobId = range jobIds {
//		//从redis中查出该job
//		if job, err = GRedis.GetSingleJob(jobId); err != nil {
//			//redis中没有对应的job
//			if job == nil {
//				//从数据库查询对应的job
//				job, _ = db.GetSingleJob(jobId)
//				job.IsDel = 1
//			} else {
//				return
//			}
//		}
//		err = db.DB.Model(job).Where("job_id = ?", jobId).Update("IsDel", 1).Error
//	}
//
//	return
//}
//
//func (db *DBManager) RecoverJob(jobIds []string) (err error) {
//	var (
//		jobId string
//		job   *common.Job
//	)
//
//	for _, jobId = range jobIds {
//		//从redis中查出该job
//		if job, err = GRedis.GetSingleJob(jobId); err != nil {
//			//redis中没有对应的job
//			if job == nil {
//				//从数据库查询对应的job
//				job, _ = db.GetSingleJob(jobId)
//				job.IsDel = 1
//			} else {
//				return
//			}
//		}
//		err = db.DB.Model(job).Where("job_id = ?", jobId).Update("IsDel", 0).Error
//	}
//
//	return
//}
