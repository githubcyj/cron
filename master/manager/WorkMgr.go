package manager

import (
	"context"
	"crontab/common"
	"crontab/master"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"time"
)

//服务发现
type WorkMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
}

var (
	GWorkMgr *WorkMgr
)

//worker服务列表
func (workMgr *WorkMgr) ListWorker() (workArr []string, err error) {
	var (
		workKey string
		getResp *clientv3.GetResponse
		kvPair  *mvccpb.KeyValue
		ip      string
	)

	//初始化
	workArr = make([]string, 0)
	workKey = common.JOB_WORKER_DIR

	if getResp, err = workMgr.kv.Get(context.TODO(), workKey, clientv3.WithPrefix()); err != nil {
		return
	}

	for _, kvPair = range getResp.Kvs {
		ip = common.ExtracWorkIp(string(kvPair.Key))
		workArr = append(workArr, ip)
	}

	return
}

func InitWorkMgr() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
	)
	config = clientv3.Config{
		Endpoints:   master.GConfig.EtcdIP,
		DialTimeout: time.Duration(master.GConfig.EtcdTimeout) * time.Millisecond,
	}

	if client, err = clientv3.New(config); err != nil {
		return
	}

	kv = clientv3.NewKV(client)

	GWorkMgr = &WorkMgr{
		client: client,
		kv:     kv,
	}
	return
}
