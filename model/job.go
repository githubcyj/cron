package model

import (
	"crontab/master/manager"
	"encoding/json"
	"errors"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"time"
)

/**
  @author 胡烨
  @version 创建时间：2019/12/2 9:48
  @todo
*/

//定时任务
type Job struct {
	Id             int    `json:"id"`
	Name           string `json:"name"`                          //任务名
	Command        string `json:"command"`                       //任务命令
	JobId          string `json:"jobId"`                         //任务唯一id
	IsDel          int    `json:"isDel" gorm:"default:'0'`       //0：任务未删除，1：任务已删除
	UpdateCount    int    `json:"updateCount" gorm:"default:'0'` //更新计数
	ExecutionCount int    `json:"executionCount"`                //执行次数
	IsFile         int    `json:"isFile"`                        //是否是文件任务
	FileId         string `json:"fileId"`                        //文件id
	Base
}

func (job *Job) BeforeSave(scope *gorm.Scope) error {
	var (
		err error
	)
	if err = scope.SetColumn("UpdateTime", time.Now()); err != nil {
		return err
	}
	return err
}

func (job *Job) BeforeCreate(scope *gorm.Scope) error {
	var (
		id  uuid.UUID
		err error
		uid string
	)
	id, _ = uuid.NewV4()
	uid = string([]rune(id.String())[:10])
	if err = scope.SetColumn("job_id", uid); err != nil {
		return err
	}

	if err = scope.SetColumn("CreateTime", time.Now()); err != nil {
		return err
	}

	if err = scope.SetColumn("UpdateTime", time.Now()); err != nil {
		return err
	}

	return nil
}

func (job *Job) SaveDB() (err error) {
	if err = manager.GDB.DB.Create(job).Error; err != nil {
		manager.GLogMgr.WriteLog("创建任务失败：" + err.Error())
	}
	return
}

func (job *Job) SaveRedis() (err error) {

	var (
		jobStr []byte
	)

	//序列化
	if jobStr, err = json.Marshal(job); err != nil {
		return
	}

	if _, err = manager.GRedis.Conn.Do("HMSET", "jobs", job.JobId, jobStr); err != nil {
		manager.GLogMgr.WriteLog("任务保存进入数据库失败：" + err.Error())
		return
	}

	return
}

func (job *Job) UpdateJobDB() (err error) {
	var (
		oldJob *Job
	)
	//从redis中查出该job
	if oldJob, err = job.GetSingleJobRedis(); err != nil {
		//redis中没有对应的job
		if oldJob == nil {
			//从数据库查询对应的job
			oldJob, _ = job.GetSingleJobDB()
			//数据库中没有对应的内容，则需要插入数据
			if oldJob == nil {
				if err = job.SaveJobDB(); err != nil {
					return
				} else {
					return nil
				}
			}
		} else {
			return
		}
	}

	//获取旧job updateCount并加1
	job.UpdateCount = oldJob.UpdateCount + 1
	job.Id = oldJob.Id

	err = manager.GDB.DB.Save(job).Error
	return err
}

//从redis中获取单个job
func (j *Job) GetSingleJobRedis() (job *Job, err error) {
	var (
		data   interface{}
		jobStr []byte
		datas  []interface{}
	)
	if data, err = manager.GRedis.Conn.Do("hmget", "jobs", j.JobId); err != nil {
		return nil, err
	}
	datas = data.([]interface{})
	jobStr = datas[0].([]byte)
	job = &Job{}
	//反序列化
	if err = json.Unmarshal(jobStr, job); err != nil {
		return nil, err
	}
	if data == nil {
		return nil, errors.New("redis没有对应job")
	}

	return job, nil
}

//从数据库中查出单个任务
func (j *Job) GetSingleJobDB() (job *Job, err error) {
	job = &Job{}
	if err = manager.GDB.DB.Where("jobId = ?", j.JobId).First(&job).Error; err != nil {
		return nil, err
	}

	return job, err
}

//任务保存进入数据库
func (job *Job) SaveJobDB() (err error) {
	if err = manager.GDB.DB.Create(job).Error; err != nil {
		manager.GLogMgr.WriteLog("插入数据失败，失败原因：" + err.Error())
		return
	}
	return
}

//更新单个job
func (job *Job) UpdateSingleJobRedis() (err error) {
	var (
		oldJob *Job
		jobStr []byte
	)

	if oldJob, err = job.GetSingleJobRedis(); err != nil {
		return
	}

	//先删除
	if _, err = manager.GRedis.Conn.Do("hdel", "jobs", job.JobId); err != nil {
		return err
	}

	oldJob.UpdateCount++

	//序列化
	if jobStr, err = json.Marshal(oldJob); err != nil {
		return
	}
	//更新
	if _, err = manager.GRedis.Conn.Do("hmset", "jobs", oldJob.JobId, jobStr); err != nil {
		return
	}

	return nil
}

//
////从redis中获取所有job
//func (r *RedisManager) GetAllJobsRedis() (jobArr []*Job, err error) {
//	var (
//		datas  []interface{}
//		job    *common.Job
//		data   interface{}
//		ok     bool
//		d      interface{}
//		jobStr []byte
//	)
//	jobArr = make([]*common.Job, 0)
//	if data, err = r.Conn.Do("HVALS", "jobs"); err != nil {
//		GLogMgr.WriteLog("从数据库中获取数据失败，失败原因：" + err.Error())
//		return jobArr, err
//	}
//	if datas, ok = data.([]interface{}); !ok {
//		return jobArr, errors.New("类型错误")
//	}
//
//	for _, d = range datas {
//		jobStr = d.([]byte)
//		job = &common.Job{}
//		//反序列化
//		if err = json.Unmarshal(jobStr, job); err != nil {
//			return jobArr, err
//		}
//		if job.IsDel != 1 {
//			jobArr = append(jobArr, job)
//		}
//	}
//
//	return jobArr, nil
//}
