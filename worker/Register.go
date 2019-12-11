package worker

import (
	"context"
	"github.com/crontab/constants"
	"go.etcd.io/etcd/clientv3"
	"net"
	"time"
)

type Register struct {
	Client *clientv3.Client
	Kv     clientv3.KV
	Lease  clientv3.Lease
}

var (
	GRegister *Register
)

func getLocalIp() (ipv4 string, err error) {
	var (
		addrs   []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet
		isIpNet bool
	)

	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}
	for _, addr = range addrs {
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			//跳过ipv6
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String()
				return
			}
		}
	}

	return
}

func (register *Register) RegisterWorker() (err error) {
	var (
		ip                 string
		registerKey        string
		cancelCtx          context.Context
		cancelFunc         context.CancelFunc
		leaseGrantResp     *clientv3.LeaseGrantResponse
		leaseId            clientv3.LeaseID
		leaseKeepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	)

	//获取本机ip
	if ip, err = getLocalIp(); err != nil {
		return
	}

	//注册的地址
	registerKey = constants.JOB_WORKER_DIR + ip

	cancelCtx, cancelFunc = context.WithCancel(context.TODO())

	//创建租约
	if leaseGrantResp, err = register.Lease.Grant(context.TODO(), 5); err != nil {
		return
	}

	leaseId = leaseGrantResp.ID

	//自动续租
	if leaseKeepAliveChan, err = register.Lease.KeepAlive(cancelCtx, leaseId); err != nil {
		return
	}

	//启动一个协程来处理自动续租返回
	go func() {
		var (
			keepResp *clientv3.LeaseKeepAliveResponse
		)
		for {
			select {
			case keepResp = <-leaseKeepAliveChan:
				if keepResp == nil { //续租失败
					goto ERR
				}
			}
		}
	ERR:
	}()

	//注册到etcd
	if _, err = register.Kv.Put(cancelCtx, registerKey, "", clientv3.WithLease(leaseId)); err != nil {
		goto RETRY
	}

	return

RETRY:
	time.Sleep(1 * time.Second)
	if cancelFunc != nil {
		cancelFunc()
	}
	return
}

func InitRegister() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
	)
	config = clientv3.Config{
		Endpoints:   GConfig.EtcdIP,
		DialTimeout: time.Duration(GConfig.EtcdTimeout) * time.Millisecond,
	}

	if client, err = clientv3.New(config); err != nil {
		return
	}

	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	GRegister = &Register{
		Client: client,
		Kv:     kv,
		Lease:  lease,
	}
	//注册服务
	go GRegister.RegisterWorker()
	return
}
