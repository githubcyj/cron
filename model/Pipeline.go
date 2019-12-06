package model

import (
	manager "crontab/master/manager"
	"encoding/json"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
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
