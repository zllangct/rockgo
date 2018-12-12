package Cluster

import (
	"errors"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/logger"
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
	Component.Base
	locker *sync.RWMutex
	nodeComponent  *NodeComponent
	Nodes         map[string]*NodeInfo
	master   *rpc.TcpClient
}

func (this *LocationComponent) GetRequire() map[*Component.Object][]reflect.Type {
	requires := make(map[*Component.Object][]reflect.Type)
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
		reflect.TypeOf(&NodeComponent{}),
	}
	return requires
}

func (this *LocationComponent) Awake() {
	err := this.Parent.Root().Find(&this.nodeComponent)
	if err != nil {
		logger.Error("find node component failed", err)
		return
	}

	//注册位置服务节点RPC服务
	service:=new(LocationService)
	service.init(this)
	err= this.nodeComponent.Register(service)
	if err != nil {
		logger.Error(err)
		return
	}
	go this.DoLocationSync()
}

//同步节点信息到位置服务组件
func (this *LocationComponent)DoLocationSync()  {
	var reply bool
	var interval = time.Duration(Config.Config.ClusterConfig.ReportInterval)
	for {
		if this.master == nil {
			var err error
			this.master,err=this.nodeComponent.GetNodeClient(Config.Config.ClusterConfig.MasterAddress)
			if err != nil {
				logger.Error(err)
				time.Sleep(time.Second * interval)
				continue
			}
		}
		this.master.Call("MasterService.NodeInfoSync","sync",&reply)
		time.Sleep(time.Millisecond * interval)
	}
}

//查询节点信息 args : "AppID:Role:SelectorType"
func (this *LocationComponent) NodeInquiry(args string,detail bool) ([]*InquiryReply, error) {
	if this.Nodes==nil {
		return nil, errors.New("this location node is waiting to sync")
	}
	return Selector(this.Nodes).Select(args,detail,this.locker)
}
