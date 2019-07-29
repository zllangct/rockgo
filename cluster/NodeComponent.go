package Cluster

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/ecs"
	"github.com/zllangct/RockGO/config"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/rpc"
	"math/rand"
	"net"
	"reflect"
	"sync"
	"time"
)

var ErrNodeOffline = errors.New("this node is offline")

type NodeComponent struct {
	ecs.ComponentBase
	locker          sync.RWMutex
	AppName         string
	localIP         string
	Addr            string
	isOnline        bool
	islocationMode  bool
	rpcClient       sync.Map    //RPC客户端集合
	rpcServer       *rpc.Server //本节点RPC Server
	serverListener  *net.TCPListener
	locationClients []*rpc.TcpClient //位置服务器集合
	locationGetter  func()
	lockers         sync.Map //[nodeid,locker]
	clientGetting   map[string]int
}

func (this *NodeComponent) GetRequire() map[*ecs.Object][]reflect.Type {
	requires := make(map[*ecs.Object][]reflect.Type)
	requires[this.Parent().Root()] = []reflect.Type{
		reflect.TypeOf(&config.ConfigComponent{}),
	}
	return requires
}

func (this *NodeComponent) Initialize() error {
	logger.Info("NodeComponent init .....")
	this.AppName = config.Config.ClusterConfig.AppName
	this.islocationMode = config.Config.ClusterConfig.IsLocationMode
	this.clientGetting = make(map[string]int)
	//开始本节点RPC服务
	err := this.StartRpcServer()
	if err != nil {
		//地址占用，立即修改配置文件
		return err
	}
	//初始化位置服务器搜索
	this.InitLocationServerGetter()
	this.locationGetter()
	logger.Info("NodeComponent initialized.")
	return nil
}

func (this *NodeComponent) Locker() *sync.RWMutex {
	return &this.locker
}

func (this *NodeComponent) IsOnline() bool {
	this.locker.RLock()
	defer this.locker.RUnlock()

	return this.isOnline
}

//获取位置服务器
func (this *NodeComponent) InitLocationServerGetter() {
	if !config.Config.ClusterConfig.IsLocationMode {
		return
	}
	locker := sync.RWMutex{}
	isGetting := false
	//保证同时只有一个在执行
	getter := func() {
		locker.RLock()
		if isGetting {
			locker.RUnlock()
			return
		}
		locker.RUnlock()

		locker.Lock()
		isGetting = true
		locker.Unlock()

		for {
			if !this.IsOnline() {
				time.Sleep(time.Second)
				continue
			}
			g, err := this.GetNodeGroupFromMaster("location")
			if err != nil || len(g.Nodes()) == 0 {
				time.Sleep(time.Second)
				continue
			}
			cs, err := g.Clients()
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			this.locker.Lock()
			this.locationClients = cs
			this.locker.Unlock()
			locker.Lock()
			isGetting = false
			locker.Unlock()
			break
		}
	}
	this.locationGetter = func() {
		go getter()
	}
	//非频繁更新
	go func() {
		for {
			time.Sleep(time.Second * 30)
			this.locationGetter()
		}
	}()
}

func (this *NodeComponent) locationBroken() {
	this.locker.Lock()
	this.locationClients = nil
	this.locker.Unlock()
	this.locationGetter()
}

//RPC服务
func (this *NodeComponent) StartRpcServer() error {
	addr, err := net.ResolveTCPAddr("tcp", config.Config.ClusterConfig.LocalAddress)
	if err != nil {
		return err
	}
	server := rpc.NewServer()
	this.serverListener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	this.rpcServer = server
	logger.Info(fmt.Sprintf("NodeComponent RPC server listening on: [ %s ]", addr.String()))
	go server.Accept(this.serverListener)
	return nil
}

func (this *NodeComponent) Register(rcvr interface{}) error {
	return this.rpcServer.Register(rcvr)
}

func (this *NodeComponent) clientCallback(event string, data ...interface{}) {
	switch event {
	case "close":
		nodeAddr := data[0].(string)
		this.rpcClient.Delete(nodeAddr)
		logger.Info(fmt.Sprintf("  disconnect to remote node: [ %s ]", nodeAddr))
	}
}

//获取节点客户端
func (this *NodeComponent) GetNodeClient(addr string) (*rpc.TcpClient, error) {
a:
	if v, ok := this.rpcClient.Load(addr); ok {
		client:=v.(*rpc.TcpClient)
		if !client.IsClosed() {
			return client, nil
		}
	}

	this.locker.Lock()
	count, ok := this.clientGetting[addr]
	if ok {
		if count > 100 {
			this.locker.Unlock()
			return nil, errors.New("init not complete")
		} else {
			this.clientGetting[addr] += 1
		}
		this.locker.Unlock()
		time.Sleep(time.Millisecond * 100)
		goto a
	} else {
		this.clientGetting[addr] = 0
		this.locker.Unlock()
	}

	defer func() {
		this.locker.Lock()
		delete(this.clientGetting, addr)
		this.locker.Unlock()
	}()

	client, err := this.ConnectToNode(addr, this.clientCallback)
	if err != nil {
		return nil, err
	}
	this.rpcClient.Store(addr, client)
	return client, nil
}

//查询并选择一个节点
func (this *NodeComponent) GetNode(role string, selectorType ...SelectorType) (*NodeID, error) {
	var nodeID *NodeID
	var err error
	//优先查询位置服务器
	if this.islocationMode {
		nodeID, err = this.GetNodeFromLocation(role, selectorType...)
		if err == nil {
			return nodeID, nil
		}
	}
	//位置服务器不存在或不可用时在master上查询
	nodeID, err = this.GetNodeFromMaster(role, selectorType...)
	if err != nil {
		return nil, err
	}
	return nodeID, nil
}

//查询获取客户端
func (this *NodeComponent) GetNodeClientByRole(role string, selectorType ...SelectorType) (*rpc.TcpClient, error) {
	nodeID, err := this.GetNode(role, selectorType...)
	if err != nil {
		return nil, err
	}
	client, err := nodeID.GetClient()
	if err != nil {
		return nil, err
	}
	return client, nil
}

//查询节点组
func (this *NodeComponent) GetNodeGroup(role string) (*NodeIDGroup, error) {
	var nodeIDGroup *NodeIDGroup
	var err error
	//优先查询位置服务器
	if this.islocationMode {
		nodeIDGroup, err = this.GetNodeGroupFromLocation(role)
		if err == nil {
			return nodeIDGroup, nil
		}
	}

	//位置服务器不存在或不可用时在master上查询
	nodeIDGroup, err = this.GetNodeGroupFromMaster(role)
	if err != nil {
		return nil, err
	}

	if nodeIDGroup == nil {
		nodeIDGroup = NewNodeIDGrop()
	}
	return nodeIDGroup, nil
}

//从位置服务器查询并选择一个节点
func (this *NodeComponent) GetNodeFromLocation(role string, selectorType ...SelectorType) (*NodeID, error) {
	var client *rpc.TcpClient
	var err error

	this.locker.RLock()
	if this.locationClients == nil {
		this.locker.RUnlock()
		return nil, errors.New("location server not found")
	}
	//随机一个节点
	rnd := rand.Intn(len(this.locationClients))
	client = this.locationClients[rnd]
	this.locker.RUnlock()

	var reply *[]*InquiryReply
	args := []string{
		SELECTOR_TYPE_DEFAULT, config.Config.ClusterConfig.AppName, role,
	}
	if len(selectorType) > 0 {
		args[0] = selectorType[0]
	}

	err = client.Call("LocationService.NodeInquiry", args, &reply)
	if err != nil {
		this.locationBroken()
		return nil, err
	}
	if len(*reply) > 0 {
		g := &NodeID{
			nodeComponent: this,
			Addr:          (*reply)[0].Node,
		}
		return g, nil
	}

	return nil, errors.New("no node of this role:" + role)
}

//从位置服务器查询
func (this *NodeComponent) GetNodeGroupFromLocation(role string) (*NodeIDGroup, error) {
	var client *rpc.TcpClient
	var err error

	this.locker.RLock()
	if this.locationClients == nil {
		this.locker.RUnlock()
		return nil, errors.New("location server not found")
	}
	//随机一个节点
	rnd := rand.Intn(len(this.locationClients))
	client = this.locationClients[rnd]
	this.locker.RUnlock()

	var reply *[]*InquiryReply
	args := []string{
		SELECTOR_TYPE_GROUP, config.Config.ClusterConfig.AppName, role,
	}
	err = client.Call("LocationService.NodeInquiry", args, &reply)
	if err != nil {
		this.locationBroken()
		return nil, err
	}
	g := &NodeIDGroup{
		nodeComponent: this,
		nodes:         *reply,
	}
	return g, nil
}

//从master查询并选择一个节点
func (this *NodeComponent) GetNodeFromMaster(role string, selectorType ...SelectorType) (*NodeID, error) {
	if !this.IsOnline() {
		return nil, ErrNodeOffline
	}
	client, err := this.GetNodeClient(config.Config.ClusterConfig.MasterAddress)
	if err != nil {
		return nil, err
	}
	var reply *[]*InquiryReply
	args := []string{
		SELECTOR_TYPE_DEFAULT, config.Config.ClusterConfig.AppName, role,
	}
	if len(selectorType) > 0 {
		args[0] = selectorType[0]
	}
	err = client.Call("MasterService.NodeInquiry", args, &reply)
	if err != nil {
		return nil, err
	}
	if len(*reply) > 0 {
		g := &NodeID{
			nodeComponent: this,
			Addr:          (*reply)[0].Node,
		}
		return g, nil
	}

	return nil, errors.New("no node of this role:" + role)
}

//从master查询
func (this *NodeComponent) GetNodeGroupFromMaster(role string) (*NodeIDGroup, error) {
	if !this.IsOnline() {
		return nil, ErrNodeOffline
	}
	client, err := this.GetNodeClient(config.Config.ClusterConfig.MasterAddress)
	if err != nil {
		return nil, err
	}
	var reply *[]*InquiryReply
	args := []string{
		SELECTOR_TYPE_GROUP, config.Config.ClusterConfig.AppName, role,
	}
	err = client.Call("MasterService.NodeInquiry", args, &reply)
	if err != nil {
		return nil, err
	}
	g := &NodeIDGroup{
		nodeComponent: this,
		nodes:         *reply,
	}
	return g, nil
}

//连接到某个节点
func (this *NodeComponent) ConnectToNode(addr string, callback func(event string, data ...interface{})) (*rpc.TcpClient, error) {
	client, err := rpc.NewTcpClient("tcp", addr, callback)
	if err != nil {
		return nil, err
	}

	this.rpcClient.Store(addr, client)

	logger.Info(fmt.Sprintf("  connect to node: [ %s ] success", addr))
	return client, nil
}
