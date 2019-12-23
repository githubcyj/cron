package model

import (
	"encoding/json"
	"github.com/crontab/master/manager"
	"github.com/jinzhu/gorm"
	"time"
)

/**
  @author 胡烨
  @version 创建时间：2019/12/20 19:04
  @todo 禁止执行的命令
*/
type CommandBan struct {
	id      int    `json:"id"`
	Command string `json:"command"`
	Base
}

func (commandBan *CommandBan) BeforeSave(scope *gorm.Scope) error {
	var (
		err error
	)
	if err = scope.SetColumn("UpdateTime", time.Now()); err != nil {
		return err
	}
	return err
}

func (commandBan *CommandBan) BeforeCreate(scope *gorm.Scope) error {
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

func (commandBan *CommandBan) SaveDB() (err error) {
	err = manager.GDB.DB.Create(commandBan).Error
	return
}

func (commandBan *CommandBan) SaveRedis() (err error) {
	var (
		commandStr []byte
	)

	//序列化
	if commandStr, err = json.Marshal(commandBan); err != nil {
		return
	}

	if _, err = manager.GRedis.Conn.Do("HSET", "banCommand", commandStr); err != nil {
		manager.GLogMgr.WriteLog("禁止命令保存至redis失败：" + err.Error())
		return
	}

	return
}

func (commandBan *CommandBan) GetCommand() (list []*CommandBan, err error) {
	var (
		cStr     interface{}
		bytes    []byte
		cInter   []interface{}
		commandR *CommandBan
	)
	list = make([]*CommandBan, 0)
	if cStr, err = manager.GRedis.Conn.Do("HVALS", "BANcOMMAND"); err != nil {
		return
	}
	cInter = cStr.([]interface{})
	for _, cStr = range cInter {
		bytes = cStr.([]byte)
		commandR = &CommandBan{}
		//反序列化
		if err = json.Unmarshal(bytes, commandR); err != nil {
			return
		}
		list = append(list, commandR)
	}
	return
}
