package Cluster

import (
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/utils"
	"reflect"
	"sync"
	"time"
)

type MasterComponent struct {
	Component.Base
	locker          *sync.RWMutex
	nodeComponent   *NodeComponent
	Nodes           map[string]*NodeInfo

	timeoutChecking map[string]*int

	config 			*Config.ConfigComponent
}

func (this *MasterComponent) GetRequire() map[*Component.Object][]reflect.Type {
	requires := make(map[*Component.Object][]reflect.Type)
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
		reflect.TypeOf(&NodeComponent{}),
	}
	return requires
}

func (this *MasterComponent) Awake() {
	this.Nodes = make(map[string]*NodeInfo)

	err:= this.Parent.Root().Find(&this.config)
	if err != nil {
		logger.Error("get config component failed")
		panic(err)
		return
	}

	err = this.Parent.Root().Find(&this.nodeComponent)
	if err != nil {
		logger.Error("find node component failed", err)
		return
	}

	//注册Master服务
	s := new(MasterService)
	s.init(this)
	_=this.nodeComponent.rpcServer.Register(s)

	if !this.config.CommonConfig.Debug || false{
		go this.TimeoutCheck()
	}
}

//上报节点信息
func (this *MasterComponent) UpdateNodeInfo(args *NodeInfo) {
	this.locker.Lock()
	this.Nodes[args.Address] = args
	*this.timeoutChecking[args.Address]=0
	this.locker.Unlock()
}

//查询节点信息 args : "AppID:Role:SelectorType"
func (this *MasterComponent) NodeInquiry(args string,detail bool) ([]*InquiryReply, error) {
	return Selector(this.Nodes).Select(args, detail,this.locker)
}

//检查超时节点
func (this *MasterComponent) TimeoutCheck() map[string]*NodeInfo {
	var interval = time.Duration(this.config.ClusterConfig.ReportInterval)
	for{
		time.Sleep(interval)
		this.locker.Lock()
		for addr, count := range this.timeoutChecking {
			*count = *count + 1
			if *count > 3 {
				delete(this.Nodes, addr)
				delete(this.timeoutChecking,addr)
			}
		}
		this.locker.Unlock()
	}
}

//深度复制节点信息
func (this *MasterComponent) NodesCopy() map[string]*NodeInfo {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return utils.Copy(this.Nodes).(map[string]*NodeInfo)
}

