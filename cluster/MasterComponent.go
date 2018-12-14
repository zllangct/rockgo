package Cluster

import (
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
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
	NodesOffline    map[string]struct{}
	timeoutChecking map[string]*int
}

func (this *MasterComponent) GetRequire() map[*Component.Object][]reflect.Type {
	requires := make(map[*Component.Object][]reflect.Type)
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
		reflect.TypeOf(&NodeComponent{}),
	}
	return requires
}

func (this *MasterComponent) Awake() error{
	this.locker = &sync.RWMutex{}
	this.Nodes = make(map[string]*NodeInfo)
	this.NodesOffline =make(map[string]struct{})
	this.timeoutChecking =make(map[string]*int)

	err := this.Parent.Root().Find(&this.nodeComponent)
	if err != nil {
		return err
	}

	//注册Master服务
	s := new(MasterService)
	s.init(this)
	err=this.nodeComponent.Register(s)
	if err != nil {
		return err
	}
	if !Config.Config.CommonConfig.Debug || false{
		go this.TimeoutCheck()
	}
	return nil
}

//上报节点信息
func (this *MasterComponent) UpdateNodeInfo(args *NodeInfo) {
	this.locker.Lock()
	this.Nodes[args.Address] = args
	delete(this.NodesOffline, args.Address)
	c:=this.timeoutChecking[args.Address]
	if c == nil {
		c=new(int)
	}
	*c=0

	this.locker.Unlock()
}

//查询节点信息 args : "AppID:Role:SelectorType"
func (this *MasterComponent) NodeInquiry(args string,detail bool) ([]*InquiryReply, error) {
	return Selector(this.Nodes).Select(args, detail,this.locker)
}

//检查超时节点
func (this *MasterComponent) TimeoutCheck() map[string]*NodeInfo {
	var interval = time.Duration(Config.Config.ClusterConfig.ReportInterval)
	for{
		time.Sleep(time.Millisecond* interval)
		this.locker.Lock()
		for addr, count := range this.timeoutChecking {
			*count = *count + 1
			if *count > 3 {
				delete(this.Nodes, addr)
				delete(this.timeoutChecking,addr)
				this.NodesOffline[addr]= struct{}{}
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
func (this *MasterComponent) NodesOfflineCopy() map[string]struct{} {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return utils.Copy(this.NodesOffline).(map[string]struct{})
}

