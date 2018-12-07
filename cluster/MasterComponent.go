package Cluster

import (
	"errors"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/utils"
	"reflect"
	"strings"
	"sync"
)

type MasterComponent struct {
	Component.Base
	locker       sync.RWMutex
	nodeComponent *NodeComponent
	Nodes         map[string]*NodeInfo
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

	err := this.Parent.Root().Find(&this.nodeComponent)
	if err != nil {
		logger.Error("find node component failed", err)
		return
	}

	//注册Master服务
	s := new(MasterService)
	s.init(this)
	this.nodeComponent.rpcServer.Register(s)

}

//上报节点信息
func (this *MasterComponent) UpdateNodeInfo(args *NodeInfo) {
	this.locker.Lock()
	this.Nodes[args.Address] = args
	this.locker.Unlock()
}

//查询节点信息 args : "AppID:Role"
func (this *MasterComponent) NodeInquiry(args string) ([]*InquiryReply, error) {
	arg := strings.Split(args, ":")
	if len(arg) != 2 {
		return nil, errors.New("query string wrong")
	}
	err := errors.New("no available node ")
	var reply []*InquiryReply
	this.locker.RLock()
	for nodeName, nodeInfo := range master.Nodes {
		for _, name := range nodeInfo.AppName {
			if name == arg[0] {
				for _, role := range nodeInfo.Group {
					if role == arg[1] {
						reply = append(reply, &InquiryReply{
							Node: nodeName,
							Info: nodeInfo.Info,
						})
						err = nil
						break
					}
				}
				break
			}
		}
	}
	this.locker.RUnlock()
	return reply, err
}

func (this *MasterComponent) NodesCopy() map[string]*NodeInfo {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return utils.Copy(this.Nodes).(map[string]*NodeInfo)
}