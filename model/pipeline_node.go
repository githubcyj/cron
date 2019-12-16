package model

import (
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"time"
)

/**
  @author 胡烨
  @version 创建时间：2019/12/16 10:55
  @todo 流水线与节点关联信息
*/

type PipelineNode struct {
	Id             int    `json:"id"`               //主键
	PipelineNodeId string `json:"pipeline_node_id"` //唯一id
	PipelineId     string `json:"pipeline_id"`      //流水线id
	NodeId         string `json:"node_id"`          //节点id
	Base
}

func (pipelineNode *PipelineNode) BeforeSave(scope *gorm.Scope) error {
	var (
		err error
	)
	if err = scope.SetColumn("UpdateTime", time.Now()); err != nil {
		return err
	}
	return err
}

func (pipelineNode *PipelineNode) BeforeCreate(scope *gorm.Scope) error {
	var (
		id  uuid.UUID
		err error
		uid string
	)
	id = uuid.NewV4()
	uid = string([]rune(id.String())[:10])
	if err = scope.SetColumn("node_id", uid); err != nil {
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
