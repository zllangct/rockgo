package Cluster

import (
	"errors"
	"github.com/zllangct/RockGO/ecs"
	"github.com/zllangct/RockGO/config"
	"github.com/zllangct/RockGO/rpc"
	"reflect"
	"sync"
	"time"
)

type LocationReply struct {
	NodeNetAddress map[string]string //[node id , ip]
}
type LocationQuery struct {
	Group  string
	AppID  string
	NodeID string
}

type LocationComponent struct {
	ecs.ComponentBase
	locker        *sync.RWMutex
	nodeComponent *NodeComponent
	Nodes         map[string]*NodeInfo
	NodeLog       *NodeLogs
	master        *rpc.TcpClient
}

func (this *LocationComponent) GetRequire() map[*ecs.Object][]reflect.Type {
	requires := make(map[*ecs.Object][]reflect.Type)
	requires[this.Runtime().Root()] = []reflect.Type{
		reflect.TypeOf(&config.ConfigComponent{}),
		reflect.TypeOf(&NodeComponent{}),
	}
	return requires
}

func (this *LocationComponent) Awake(ctx *ecs.Context) {
	this.locker = &sync.RWMutex{}
	err := this.Parent().Root().Find(&this.nodeComponent)
	if err != nil {
		panic(err)
	}

	//注册位置服务节点RPC服务
	service := new(LocationService)
	service.init(this)
	err = this.nodeComponent.Register(service)
	if err != nil {
		panic(err)
	}
	go this.DoLocationSync()
}

//同步节点信息到位置服务组件
func (this *LocationComponent) DoLocationSync() {
	var reply *NodeInfoSyncReply
	var interval = time.Duration(config.Config.ClusterConfig.LocationSyncInterval)
	for {
		if this.master == nil {
			var err error
			this.master, err = this.nodeComponent.GetNodeClient(config.Config.ClusterConfig.MasterAddress)
			if err != nil {
				time.Sleep(time.Second * interval)
				continue
			}
		}
		err := this.master.Call("MasterService.NodeInfoSync", "sync", &reply)
		if err != nil {
			this.master = nil
			continue
		}
		this.locker.Lock()
		this.Nodes = reply.Nodes
		this.NodeLog = reply.NodeLog
		this.locker.Unlock()
		time.Sleep(time.Millisecond * interval)
	}
}

//查询节点信息 args : "AppID:Role:SelectorType"
func (this *LocationComponent) NodeInquiry(args []string, detail bool) ([]*InquiryReply, error) {
	if this.Nodes == nil {
		return nil, errors.New("this location node is waiting to sync")
	}
	return Selector(this.Nodes).DoQuery(args, detail, this.locker)
}

//日志获取
func (this *LocationComponent) NodeLogInquiry(args int64) ([]*NodeLog, error) {
	this.locker.RLock()
	defer this.locker.RUnlock()

	if this.NodeLog == nil {
		return nil, errors.New("this location node is waiting to sync")
	}
	return this.NodeLog.Get(args), nil
}
