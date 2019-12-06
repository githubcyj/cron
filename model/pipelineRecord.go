package model

import (
	"github.com/jinzhu/gorm"
	"time"
)

/**
  @author 胡烨
  @version 创建时间：2019/12/6 16:59
  @todo 流水线记录
*/
type PipelineRecord struct {
	Id           string    //流水线记录唯一id,uuid生成
	PipelineId   string    //流水线号
	PipelineName string    //流水线名称
	Type         int       //流水线类型 定时任务 0 延时任务 1
	Spec         string    //定时器
	Status       int       //状态
	Duration     int       //持续时间
	BeginWith    time.Time //开始于
	FinishWith   time.Time //结束于
	Base
}

func (pipelineRecord *PipelineRecord) BeforeCreate(scope *gorm.Scope) error {
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
