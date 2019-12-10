package model

import (
	"context"
	"crontab/constants"
	manager "crontab/master/manager"
	"encoding/json"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"go.etcd.io/etcd/clientv3"
	"strconv"
	"time"
)

/**
  @author 胡烨
  @version 创建时间：2019/11/29 14:11
  @todo 流水线model
*/

//流水线
type Pipeline struct {
	Id             int    `json:"id"`
	PipelineId     string `json:"pipelineId"`                    //流水线id
	Status         int    `json:"status"`                        //流水线状态
	Name           string `json:"name"`                          //名称
	Type           int    `json:"type"`                          //流水线类型 0：定时执行，1：延时执行
	CronExpr       string `json:"cronExpr"`                      //定时任务：定时任务执行时间
	IsDel          int    `json:"isDel" gorm:"default:'0'`       //0：流水线未删除，1：流水线已删除
	UpdateCount    int    `json:"updateCount" gorm:"default:'0'` //更新计数
	TimerExecuter  string `json:"timerExecuter"`                 //延时任务执行时间
	ExecutionCount int    `json:"executionCount"`                //执行次数
	Finished       string `json:"finished"`                      //成功时执行的任务id
	Failed         string `json:"failed"`                        //失败时执行的任务id
	FinishedJob    *Job
	FailedJob      *Job
	Steps          []*PipelineJob
	Base
}

func (pipeline *Pipeline) BeforeSave(scope *gorm.Scope) error {
	var (
		err error
	)
	if err = scope.SetColumn("UpdateTime", time.Now()); err != nil {
		return err
	}
	return err
}

func (pipeline *Pipeline) BeforeCreate(scope *gorm.Scope) error {
	var (
		id  uuid.UUID
		err error
		uid string
	)
	id, _ = uuid.NewV4()
	uid = string([]rune(id.String())[:10])
	if err = scope.SetColumn("pipelineId", uid); err != nil {
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

func (pipeline *Pipeline) SaveDB() (err error) {
	if err = manager.GDB.DB.Create(pipeline).Error; err != nil {
		manager.GLogMgr.WriteLog("创建流水线失败：" + err.Error())
	}
	return
}

func (pipeline *Pipeline) SaveRedis() (err error) {

	var (
		pipeStr []byte
	)

	//序列化
	if pipeStr, err = json.Marshal(pipeline); err != nil {
		return
	}

	if _, err = manager.GRedis.Conn.Do("HMSET", "pipeline", pipeline.PipelineId, pipeStr); err != nil {
		manager.GLogMgr.WriteLog("流水线保存进入数据库失败：" + err.Error())
		return
	}

	return
}

func (p *Pipeline) GetRedis() (pipeline *Pipeline, err error) {
	var (
		pStr   interface{}
		bytes  []byte
		pInter []interface{}
	)

	pipeline = &Pipeline{}

	if pStr, err = manager.GRedis.Conn.Do("HMGET", "pipeline", p.PipelineId); err != nil {
		manager.GLogMgr.WriteLog("从redis中获取流水线失败:" + err.Error())
		return nil, err
	}
	pInter = pStr.([]interface{})
	bytes = pInter[0].([]byte)
	//反序列化
	if err = json.Unmarshal(bytes, pipeline); err != nil {
		manager.GLogMgr.WriteLog("反序列化流水线失败：" + err.Error())
		return nil, err
	}

	return pipeline, nil
}

func (p *Pipeline) DelRedis() (err error) {
	_, err = manager.GRedis.Conn.Do("HDEL", "pipeline", p.PipelineId)
	return
}

func (p *Pipeline) DelDB() (err error) {
	if err = manager.GDB.DB.Where("pipeline_id=?", p.PipelineId).Delete(&Pipeline{}).Error; err != nil {
		manager.GLogMgr.WriteLog("删除流水线失败：" + err.Error())
	}
	return
}

func (p *Pipeline) SaveEtcd() (err error) {
	var (
		saveKey     string
		putResponse *clientv3.PutResponse
		old         *PipelineJob
		pipeStr     []byte
	)

	//保存的路径
	saveKey = constants.SAVE_JOB_DIR + p.PipelineId

	//序列化
	if pipeStr, err = json.Marshal(p); err != nil {
		manager.GLogMgr.WriteLog("流水线序列化失败：" + err.Error())
		return
	}

	//需要传回上一次保存的数据，用以反馈
	if putResponse, err = manager.GJobMgr.Kv.Put(context.TODO(), saveKey, string(pipeStr), clientv3.WithPrevKV()); err != nil {
		return
	}
	manager.GLogMgr.WriteLog("添加任务成功,任务：" + p.Name)

	//如果是更新，则需要返回旧值
	if putResponse.PrevKv != nil {
		//反序列化
		_ = json.Unmarshal(putResponse.PrevKv.Value, old)
		manager.GLogMgr.WriteLog("更新操作")
	}
	return
}

//强杀任务
func (p *Pipeline) KillEtcd() (err error) {

	var (
		killKey        string
		putResp        *clientv3.PutResponse
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId        clientv3.LeaseID
	)

	killKey = constants.KILL_JOB_DIR + p.PipelineId

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

//将任务推送到消息队列
func (pipeline *Pipeline) PushMq() (err error) {
	var (
		pipelineStr []byte
		now         time.Time
		diff        time.Duration
	)
	//序列化
	if pipelineStr, err = json.Marshal(pipeline); err != nil {
		return
	}
	//设置过期时间
	loc, _ := time.LoadLocation("Local")
	the_time, _ := time.ParseInLocation("2006-01-02 15:04:05", pipeline.TimerExecuter, loc)

	//获取当前时间
	now = time.Now()
	//与当前时间的差
	diff = now.Sub(the_time)

	if err = manager.GMqMgrProduce.Ch.Publish(
		constants.DELAY_EXCHANGE,
		constants.DELAY_KEY,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        pipelineStr,
			Expiration:  strconv.FormatFloat(diff.Seconds(), 'E', -1, 64), //设置过期时间
		},
	); err != nil {
		return
	}
	return
}
