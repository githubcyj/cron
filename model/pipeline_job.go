package model

import (
	"context"
	"crontab/common"
	"crontab/master/manager"
	"encoding/json"
	"errors"
	"github.com/jinzhu/gorm"
	"go.etcd.io/etcd/clientv3"
	"time"
)

/**
  @author 胡烨
  @version 创建时间：2019/11/29 14:15
  @todo 流水线与任务关联模型
*/
type PipelineJob struct {
	Id         int    `json:"id"`         //主键id
	PipelineId string `json:"pipelineId"` //流水线id
	JobId      string `json:"jobId"`      //任务id
	Step       int    `json:"step"`       //步骤
	Timeout    int    `json:"timeout"`    //超时时间
	Interval   int    `json:"interval"`   //间隔时间
	Retries    int    `json:"retries"`    //重试次数
	Job        *Job
	Base
}

func (pipelineJob *PipelineJob) BeforeCreate(scope *gorm.Scope) error {
	var (
		err error
	)
	if err = scope.SetColumn("CreateTime", time.Now()); err != nil {
		return err
	}

	if err = scope.SetColumn("UpdateTime", time.Now()); err != nil {
		return err
	}

	return nil
}

func (pipelineJob *PipelineJob) SaveDB() (err error) {
	if err = manager.GDB.DB.Create(pipelineJob).Error; err != nil {
		manager.GLogMgr.WriteLog("创建流水线失败：" + err.Error())
	}
	return
}

func (pipelineJob *PipelineJob) SaveRedis() (err error) {

	var (
		data interface{}
		len  int
		pStr []byte
	)

	//先获取任务个数
	if data, err = manager.GRedis.Conn.Do("ZCARD", pipelineJob.PipelineId); err != nil {
		manager.GLogMgr.WriteLog("获取流水线下任务个数失败：" + err.Error())
	}

	len = int(data.(int64)) + 1

	//序列化
	if pStr, err = json.Marshal(pipelineJob); err != nil {
		return
	}

	if _, err = manager.GRedis.Conn.Do("ZADD", pipelineJob.PipelineId, len, pStr); err != nil {
		manager.GLogMgr.WriteLog("流水线任务关系保存redis失败：" + err.Error())
		return
	}

	return
}

func (pipelineJob *PipelineJob) DelRedis() (err error) {
	if _, err = manager.GRedis.Conn.Do("DEL", pipelineJob.PipelineId); err != nil {
		manager.GLogMgr.WriteLog("删除列表失败：" + err.Error())
	}

	return
}

func (pipelineJob *PipelineJob) GetPipelineJobLen() (len int, err error) {
	var (
		length int
		data   interface{}
	)

	if data, err = manager.GRedis.Conn.Do("ZCARD", pipelineJob.PipelineId); err != nil {
		manager.GLogMgr.WriteLog("获取流水线下任务个数失败：" + err.Error())
		return length, err
	}

	length = int(data.(int64))
	if length == 0 {
		//从数据库获取
		if err = manager.GDB.DB.Model(&PipelineJob{}).Where("pipeline_id=?", pipelineJob.PipelineId).Count(&length).Error; err != nil {
			return length, err
		}
	}

	return length, nil
}

func (pipelineJob *PipelineJob) GetAllJobRedis() (jobArr []*PipelineJob, err error) {
	var (
		datas  []interface{}
		job    *PipelineJob
		data   interface{}
		ok     bool
		d      interface{}
		jobStr []byte
	)

	jobArr = make([]*PipelineJob, 0)
	if data, err = manager.GRedis.Conn.Do("ZRANGE", pipelineJob.PipelineId, 0, -1); err != nil {
		manager.GLogMgr.WriteLog("从数据库中获取数据失败，失败原因：" + err.Error())
		return jobArr, err
	}
	if datas, ok = data.([]interface{}); !ok {
		return jobArr, errors.New("类型错误")
	}

	for _, d = range datas {
		jobStr = d.([]byte)
		job = &PipelineJob{}
		//反序列化
		if err = json.Unmarshal(jobStr, job); err != nil {
			return jobArr, err
		}
		jobArr = append(jobArr, job)
	}

	return jobArr, nil
}

func (pipelineJob *PipelineJob) SaveEtcd() (err error) {
	var (
		saveKey     string
		putResponse *clientv3.PutResponse
		old         *PipelineJob
		pipeStr     []byte
	)

	//保存的路径
	saveKey = common.SAVE_JOB_DIR + pipelineJob.PipelineId

	//序列化
	if pipeStr, err = json.Marshal(pipelineJob); err != nil {
		manager.GLogMgr.WriteLog("流水线序列化失败：" + err.Error())
		return
	}

	//需要传回上一次保存的数据，用以反馈
	if putResponse, err = manager.GJobMgr.Kv.Put(context.TODO(), saveKey, string(pipeStr), clientv3.WithPrevKV()); err != nil {
		return
	}
	manager.GLogMgr.WriteLog("添加任务成功,任务：" + pipelineJob.JobId)

	//如果是更新，则需要返回旧值
	if putResponse.PrevKv != nil {
		//反序列化
		_ = json.Unmarshal(putResponse.PrevKv.Value, old)
		manager.GLogMgr.WriteLog("更新操作")
	}
	return
}

//强杀任务
func (pipelineJob *PipelineJob) KillEtcd() (err error) {

	var (
		killKey        string
		putResp        *clientv3.PutResponse
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId        clientv3.LeaseID
	)

	killKey = common.KILL_JOB_DIR + pipelineJob.PipelineId

	//让workder监听到一次put操作，创建一个租约让其稍后自动过期
	if leaseGrantResp, err = manager.GJobMgr.Lease.Grant(context.TODO(), 1); err != nil {
		return
	}

	leaseId = leaseGrantResp.ID

	if putResp, err = manager.GJobMgr.Kv.Put(context.TODO(), killKey, "", clientv3.WithLease(leaseId)); err != nil {
		return
	}

	if putResp.PrevKv != nil {
		//反序列化json
		return
	}
	return
}

//任务是否与流水线绑定
func (pipelineJob *PipelineJob) IsJobBindPipeline() (bind bool, err error) {
	var (
		count int
	)
	if err = manager.GDB.DB.Model(&PipelineJob{}).Where("job_id=?", pipelineJob.JobId).Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	} else {
		return false, nil
	}
}
