package model

import (
	"crontab/master/manager"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"time"
)

/**
  @author 胡烨
  @version 创建时间：2019/12/9 16:03
  @todo 文件结构体
*/

type File struct {
	Id     int
	FileId string `json:"file_id"`
	Name   string
	Base
}

func (file *File) BeforeSave(scope *gorm.Scope) error {
	var (
		err error
	)
	if err = scope.SetColumn("UpdateTime", time.Now()); err != nil {
		return err
	}
	return err
}

func (file *File) BeforeCreate(scope *gorm.Scope) error {
	var (
		id  uuid.UUID
		err error
		uid string
	)
	id, _ = uuid.NewV4()
	uid = string([]rune(id.String())[:10])
	if err = scope.SetColumn("file_id", uid); err != nil {
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

func (f *File) AfterCreate(scope *gorm.Scope) error {

	var (
		err error
	)
	fmt.Println(f)
	fmt.Println(scope)
	fmt.Println(err)

	return err
}

//保存进入数据库
func (f *File) SaveDb() (fileId string, err error) {
	var (
		row *sql.Rows
	)
	if row, err = manager.GDB.DB.Create(f).Select("file_id").Rows(); err != nil {
		return
	}
	if row.Next() {
		row.Scan(&fileId)
	}
	return
}

func (file *File) SaveRedis() (err error) {
	_, err = manager.GRedis.Conn.Do("HMSET", "file", file.FileId, file.FileId)

	return
}

func (file *File) GetSingleRedis() (f *File, err error) {
	var (
		data  interface{}
		datas []interface{}
		fstr  []byte
	)
	f = &File{}
	if data, err = manager.GRedis.Conn.Do("HMGET", "file", file.FileId); err != nil {
		return nil, err
	}
	datas = data.([]interface{})
	if datas[0] != nil {
		fstr = datas[0].([]byte)
		if fstr != nil {
			f.FileId = string(fstr)
		} else {
			return nil, errors.New("redis没有对应job")
		}
	} else {
		return nil, errors.New("redis没有对应job")
	}
	return
}

func (file *File) GetSingleDB() (f *File, err error) {
	f = &File{}
	if err = manager.GDB.DB.Where("file_id = ?", f.FileId).First(&f).Error; err != nil {
		return nil, err
	}

	return f, err
}

func (file *File) IsExist() bool {
	var (
		fileR *File
		err   error
	)
	if fileR, err = file.GetSingleRedis(); err != nil {
		return false
	}

	if fileR == nil {
		if fileR, err = file.GetSingleDB(); err != nil {
			return false
		}
		if fileR == nil {
			return false
		}
	}
	return true
}
