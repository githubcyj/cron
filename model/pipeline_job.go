package model

import (
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
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
	Base
}

func (pipelineJob *PipelineJob) BeforeCreate(scope *gorm.Scope) error {
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
