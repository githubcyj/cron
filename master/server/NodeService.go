package server

import "github.com/crontab/util"

/**
  @author 胡烨
  @version 创建时间：2019/12/16 11:19
  @todo
*/

//服务启动时，初始化该服务节点信息
func InitNode() (err error) {
	var (
		ip string
	)
	//获得本机ip
	if ip, err = util.GetLocalIp(); err != nil {
		return
	}

	//将该ip地址存储起来，标记为master节点

	return
}
