package model

import "time"

/**
  @author 胡烨
  @version 创建时间：2019/11/29 14:12
  @todo 基础model
*/

type Base struct {
	CreateUserId string    `json:"createUserId"`
	CreateTime   time.Time `json:"createTime"`
	UpdateUserId string    `json:"updateUserId"`
	UpdateTime   time.Time `json:"updateTime"`
}
