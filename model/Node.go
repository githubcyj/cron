package model

import (
	"encoding/json"
	"github.com/crontab/constants"
	"github.com/crontab/master/manager"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"time"
)

/**
  @author 胡烨
  @version 创建时间：2019/12/16 10:45
  @todo 节点信息
*/

type Node struct {
	Id     int    `json:"id"`                       //主键id
	NodeId string `json:"node_id"`                  //唯一id
	Host   string `json:"host"`                     //主机地址
	Port   int    `json:"port"  gorm:"default:'81'` //端口
	Status int    `json:"status"`                   //状态
	Base
}

func (node *Node) BeforeSave(scope *gorm.Scope) error {
	var (
		err error
	)
	if err = scope.SetColumn("UpdateTime", time.Now()); err != nil {
		return err
	}
	return err
}

func (node *Node) BeforeCreate(scope *gorm.Scope) error {
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

func (node *Node) SaveDB() (err error) {
	if err = manager.GDB.DB.Create(node).Error; err != nil {
		manager.GLogMgr.WriteLog("ch：" + err.Error())
	}
	return
}

func (node *Node) SaveRedis() (err error) {

	var (
		pipeStr []byte
	)

	//序列化
	if pipeStr, err = json.Marshal(node); err != nil {
		return
	}

	if _, err = manager.GRedis.Conn.Do("HMSET", "node", node.Host, pipeStr); err != nil {
		manager.GLogMgr.WriteLog("节点保存进入数据库失败：" + err.Error())
		return
	}

	return
}

func (n *Node) GetNode() (node *Node, err error) {
	var (
		data    interface{}
		datas   []interface{}
		nodeStr []byte
	)
	//从缓存读取
	if data, err = manager.GRedis.Conn.Do("HMGET", "node", n.Host); err != nil {
		return nil, err
	}
	datas = data.([]interface{})
	datas = data.([]interface{})
	if datas[0] != nil {
		nodeStr = datas[0].([]byte)
		node = &Node{}
		//反序列化
		if err = json.Unmarshal(nodeStr, node); err != nil {
			return nil, err
		}
	} else {
		if err = manager.GDB.DB.Where("host = ?", n.Host).Find(node).Error; err != nil {
			return nil, err
		}
	}
	return node, nil
}

func (node *Node) GetNodeByIp() (err error) {

}

//节点下线
func (node *Node) Offline() (err error) {
	var (
		nodeR   *Node
		nodeStr []byte
	)
	node.Status = constants.NODE_ONLINE

	if nodeR, err = node.GetNode(); err != nil {
		return
	}

	//1. 删除缓存中的数据
	if _, err = manager.GRedis.Conn.Do("HDEL", "node", node.Host); err != nil {
		return
	}

	if err = manager.GDB.DB.Model(&Node{}).Where("node_id = ?", node.NodeId).Update("status", node.Status).Error; err != nil {
		return
	}
	nodeR.Status = constants.NODE_OFFLINE
	//序列化
	if nodeStr, err = json.Marshal(nodeR); err != nil {
		return
	}
	//重新添加缓存
	if _, err = manager.GRedis.Conn.Do("HMSET", "node", nodeR.Host, nodeStr); err != nil {
		return
	}

	return
}
