package manager

import (
	"fmt"
	"github.com/crontab/master/config"
	"github.com/garyburd/redigo/redis"
	"time"
)

/**
  @author 胡烨
  @version 创建时间：2019/11/18 14:46
  @todo
*/

type RedisManager struct {
	RedisPool *redis.Pool
	Conn      redis.Conn
}

var GRedis *RedisManager

func InitRedis() {
	var (
		pool *redis.Pool
		conn redis.Conn
	)
	pool = &redis.Pool{
		Dial: func() (conn redis.Conn, e error) {
			var (
				url string
			)
			url = fmt.Sprintf("%s:%d", config.GConfig.RedisHost, config.GConfig.RedisPort)
			c, err := redis.Dial("tcp", url, redis.DialPassword((config.GConfig.RedisPassword)))
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		MaxIdle:     config.GConfig.RedisMaxIdle,
		MaxActive:   config.GConfig.RedisMaxActive,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
	}
	conn = pool.Get()
	GRedis = &RedisManager{
		RedisPool: pool,
		Conn:      conn,
	}
}

//
////添加单个job到redis中
//func (r *RedisManager) AddJob(job *common.Job) (err error) {
//
//	var (
//		jobArr []byte
//	)
//	//序列化json
//	if jobArr, err = json.Marshal(job); err != nil {
//		return
//	}
//
//	if _, err = r.Conn.Do("HMSET", "jobs", job.JobId, jobArr); err != nil {
//		return err
//	}
//	return
//}
//
////从redis中获取所有job
//func (r *RedisManager) GetAllJobs() (jobArr []*common.Job, err error) {
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
//
////从redis中获取所有job
//func (r *RedisManager) GetAllDeleteJobs() (jobArr []*common.Job, err error) {
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
//		if job.IsDel == 1 {
//			jobArr = append(jobArr, job)
//		}
//	}
//
//	return jobArr, nil
//}
//
////保存所有的job到redis中
//func (r *RedisManager) AddAllJob(jobs []*common.Job) {
//	var (
//		job  *common.Job
//		err  error
//		data []byte
//	)
//
//	//将所有的任务放进redi中
//	for _, job = range jobs {
//		//序列化
//		data, _ = json.Marshal(job)
//		if _, err = r.Conn.Do("HMSET", "jobs", job.JobId, data); err != nil {
//			continue
//		}
//	}
//}

//
////更新单个job
//func (r *RedisManager) UpdateSingleJob(job *common.Job) (err error) {
//	var (
//		oldJob *common.Job
//		jobStr []byte
//	)
//
//	if oldJob, err = r.GetSingleJob(job.JobId); err != nil {
//		return
//	}
//
//	//先删除
//	if _, err = r.Conn.Do("hdel", "jobs", job.JobId); err != nil {
//		return err
//	}
//
//	oldJob.UpdateCount++
//
//	//序列化
//	if jobStr, err = json.Marshal(oldJob); err != nil {
//		return
//	}
//	//更新
//	if _, err = r.Conn.Do("hmset", "jobs", oldJob.JobId, jobStr); err != nil {
//		return
//	}
//
//	return nil
//}
//
////先删除，然后再更新，直接更新开销太大
//func (r *RedisManager) UpdateJob(jobs []*common.Job) (err error) {
//	var (
//		ids []string
//		job *common.Job
//	)
//	ids = make([]string, 0)
//	for _, job = range jobs {
//		ids = append(ids, job.JobId)
//	}
//
//	if err = r.DelJobs(ids); err != nil {
//		return
//	}
//
//	//插入
//	for _, job = range jobs {
//		if err = r.AddJob(job); err != nil {
//			return
//		}
//	}
//
//	return
//}
