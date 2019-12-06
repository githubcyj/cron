package model

import (
	"github.com/jinzhu/gorm"
	"time"
)

/**
  @author 胡烨
  @version 创建时间：2019/12/6 17:04
  @todo 任务记录
*/

type JobRecord struct {
	id               int
	PipelineRecordId string    //流水线记录id
	JobId            string    //任务id
	JobName          string    //任务名称
	Content          string    //执行内容
	Timeout          int       //超时时间
	Retries          int       //重试次数
	Status           int       //任务状态 0:失败，1：成功
	Result           string    //执行结果
	Duration         int       //持续时间
	BeginWith        time.Time //开始于
	FinishWith       time.Time //结束于
	Base
}

func (jobRecord *JobRecord) BeforeCreate(scope *gorm.Scope) error {
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
