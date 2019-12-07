package worker

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

/**
  @author 胡烨
  @version 创建时间：2019/11/18 11:27
  @todo 数据库连接
*/
/**
数据库连接相关
*/

type DBManager struct {
	DB *gorm.DB
}

var GDB *DBManager

//初始化数据库连接
func InitDB() (err error) {
	var (
		url string
		db  *gorm.DB
	)
	url = fmt.Sprintf("%s:%s@(%s:%d)/%s?allowNativePasswords=true&parseTime=True&loc=Local",
		GConfig.User, GConfig.Password, GConfig.Host, GConfig.Port, GConfig.Database)
	if db, err = gorm.Open(GConfig.Dialect, url); err != nil {
		return
	}
	db.DB().SetMaxIdleConns(GConfig.MaxIdleConns)
	db.DB().SetMaxOpenConns(GConfig.MaxOpenConns)
	GDB = &DBManager{DB: db}
	return
}
