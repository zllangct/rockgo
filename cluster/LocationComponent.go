package Cluster

import (
	"errors"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/rpc"
	"reflect"
	"strings"
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
	locker sync.RWMutex
	nodeComponent  *NodeComponent
	Nodes         map[string]*NodeInfo
	config	 *Config.ConfigComponent
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
	this.Nodes = make(map[string]*NodeInfo)
	err:= this.Parent.Root().Find(&this.config)
	if err != nil {
		logger.Fatal("get config component failed")
		panic(err)
		return
	}
	err = this.Parent.Root().Find(&this.nodeComponent)
	if err != nil {
		logger.Error("find node component failed", err)
		return
	}

	//注册位置服务节点RPC服务
	service:=new(LocationService)
	service.init(this)
	this.nodeComponent.rpcServer.Register(service)

	go this.DoLocationSync()
}

//同步节点信息到位置服务组件
func (this *LocationComponent)DoLocationSync()  {
	var reply bool
	var interval = time.Duration(this.config.ClusterConfig.ReportInterval)
	for {
		if this.master == nil {
			var err error
			this.master,err=this.nodeComponent.GetNodeClient(this.config.ClusterConfig.MasterAddress)
			if err != nil {
				logger.Error(err)
				time.Sleep(time.Second * interval)
				continue
			}
		}
		this.master.Call("MasterService.NodeInfoSynchronous","sync",&reply)
		time.Sleep(time.Second * interval)
	}
}

//选择一个节点客户端
func (this *LocationComponent) SelectNodeClient(appID string,role string,selector Selector) *rpc.TcpClient{

	return nil
}

//筛选节点客户端组
func (this *LocationComponent) SelectNodeGroup(appID string,role string) []*rpc.TcpClient{

	return nil
}

//查询节点信息 args : "AppID:Role"
func (this *LocationComponent) NodeInquiry(args string) ([]*InquiryReply, error) {
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