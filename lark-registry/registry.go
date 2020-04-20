package larkRegistry

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/zeast/logs"
	doveclientCli "gitlab.int.jumei.com/JMArch/go-doveclient-cli"
	larkProto "gitlab.int.jumei.com/JMArch/go-rpc/lark-proto"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Registry struct {
	larkNetwork string //和 lark 通信用
	larkNetaddr string
	regInfo     larkProto.ServiceRegistInfo
	sync.Mutex
	doveclientStr     string
	registerBytes     []byte
	registryCenter    map[string][]string //已经注册的注册中心的信息
	allRegistryCenter map[string][]string //所有注册中心的信息
	status            int32               // 0 is closed, 1 is running
	closeCh           chan struct{}
}

func NewRegistry(i larkProto.ServiceRegistInfo) *Registry {
	return &Registry{
		larkNetwork:   "tcp4",
		larkNetaddr:   "127.0.0.1:12312",
		regInfo:       i,
		doveclientStr: "unix:////var/lib/doveclient/doveclient.sock",
		closeCh:       make(chan struct{}),
	}
}

func (r *Registry) Register(dc []string) error {
	//获取本机的 dove 环境名称
	var envkey = "RpcPool.ENV"
	b, err := doveclientCli.NewDoveClient(r.doveclientStr).Get(envkey)
	if err != nil {
		return errors.WithMessagef(err, "can't get key:%s from dove client", envkey)
	}

	var env string
	if err := json.Unmarshal(b, &env); err != nil {
		return errors.WithMessagef(err, "can't unmarshal local env. %q", b)
	}

	//获取各个机房的信息
	var idckey = "RpcPool.Etcd.Idc2EtcdServers"
	b, err = doveclientCli.NewDoveClient(r.doveclientStr).Get(idckey)
	var rcInfo = make(map[string][]string)
	err = json.Unmarshal(b, &rcInfo)
	if err != nil {
		return errors.WithMessagef(err, "can't get key:%s from dove client", idckey)
	}

	var info = make(map[string][]string)
	if len(dc) == 0 {
		//如果不指定注册中心，那么默认注册本机中心
		i, ok := rcInfo[env]
		if !ok {
			return errors.Errorf("can't get register center info, env:%s, registryCenter:%+v", env, rcInfo)
		}
		info[env] = i
	} else {
		//如果指定了注册中心，那么选择注册中心注册
		for _, e := range dc {
			i, ok := rcInfo[e]
			if !ok {
				return errors.Errorf("can't get register center info, env:%s, registryCenter:%+v", e, rcInfo)
			}
			info[e] = i
		}
	}

	registerBytes, _ := larkProto.PackRegisterReq([]larkProto.ServiceRegistInfo{r.regInfo}, info)
	r.registryCenter = info
	r.allRegistryCenter = rcInfo

	r.Lock()
	defer r.Unlock()

	r.registerBytes = registerBytes

	go r.run()

	return nil
}

func (r *Registry) run() {
	if !atomic.CompareAndSwapInt32(&r.status, 0, 1) {
		return
	}

	r.Lock()
	r.send(r.registerBytes)
	r.Unlock()

	tk := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-tk.C:
			r.Lock()
			r.send(r.registerBytes)
			r.Unlock()

		case <-r.closeCh:
			return
		}
	}

}

func (r *Registry) UnRegister(registerCenterList []string) {
	if atomic.LoadInt32(&r.status) == 0 {
		return
	}

	r.Lock()
	defer r.Unlock()

	var rcInfo map[string][]string
	if len(registerCenterList) == 0 {
		//默认全部注销
		rcInfo = r.registryCenter
	} else {
		//按照参数注销
		rcInfo = make(map[string][]string)
		for _, registerCenter := range registerCenterList {
			if ipList, ok := r.registryCenter[registerCenter]; ok {
				rcInfo[registerCenter] = ipList
				delete(r.registryCenter, registerCenter)
			}
		}
	}

	r.registerBytes, _ = larkProto.PackRegisterReq([]larkProto.ServiceRegistInfo{r.regInfo}, r.registryCenter)
	b, _ := larkProto.PackUnRegisterReq([]larkProto.ServiceRegistInfo{r.regInfo}, rcInfo)
	r.send(b)
}

func (r *Registry) Close() {
	if atomic.CompareAndSwapInt32(&r.status, 1, 0) {
		close(r.closeCh)
	}
}

func (r *Registry) RegistryCenter() map[string][]string {
	return r.registryCenter
}

func (r *Registry) AllRegistryCenter() map[string][]string {
	return r.allRegistryCenter
}

func (r *Registry) send(b []byte) {
	conn, err := net.Dial(r.larkNetwork, r.larkNetaddr)
	if err != nil {
		logs.Error(err)
		return
	}

	defer conn.Close()

	_, err = conn.Write(b)
	if err != nil {
		logs.Error(err)
		return
	}

}
