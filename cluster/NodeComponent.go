package Cluster

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/rpc"
	"net"
	"reflect"
	"sync"
	"time"
)


var ErrNodeOffline =errors.New("this node is offline")

type NodeComponent struct {
	Component.Base
	locker sync.RWMutex
	AppName         string
	localIP         string
	Addr string
	isOnline       	bool
	islocationMode  bool
	rpcClient      	sync.Map 				//RPC客户端集合
	rpcServer      	*rpc.Server				//本节点RPC Server
	serverListener   *net.TCPListener
 	locationClients *NodeIDGroup			//位置服务器集合
	lockers      	sync.Map 				//[nodeid,locker]
}

func (this *NodeComponent) GetRequire() (map[*Component.Object][]reflect.Type) {
	requires:=make(map[*Component.Object][]reflect.Type)
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
	}
	return requires
}

func(this *NodeComponent)Awake()error{
	this.AppName = Config.Config.ClusterConfig.AppName
	this.islocationMode =Config.Config.ClusterConfig.IsLocationMode
	//开始本节点RPC服务
	err:= this.StartRpcServer()
	if err!=nil {
		return err
	}
	//查询位置服务器
	go this.GetLocationServer()
	return nil
}

//网络模块不能立即做清理，其他模块清理过程中会进行通讯
//func (this *NodeComponent)Destroy()error {
//	err:= this.serverListener.Close()
//	this.rpcClient.Range(func(key, value interface{}) bool {
//		err=value.(*rpc.TcpClient).Close()
//		if err!=nil {
//			logger.Error(err)
//		}
//		return true
//	})
//	return err
//}

func (this *NodeComponent)Locker() *sync.RWMutex {
	return &this.locker
}

func (this *NodeComponent)IsOnline() bool {
	this.locker.RLock()
	defer this.locker.RUnlock()

	return this.isOnline
}

//获取位置服务器
func (this *NodeComponent)GetLocationServer()  {
	for {
		if !this.IsOnline() {
			time.Sleep(time.Second)
			continue
		}
		this.locker.Lock()
		if this.locationClients != nil {
			this.locker.Unlock()
			time.Sleep(time.Second)
			continue
		}
		this.locker.Unlock()
		g,err:= this.GetNodeGroupFromMaster("location")
		if err!=nil || len(g.Nodes())==0{
			//logger.Debug(err)
			time.Sleep(time.Second)
			continue
		}
		this.locker.Lock()
		this.locationClients = g
		this.locker.Unlock()
		continue
	}
}

//RPC服务
func (this *NodeComponent) StartRpcServer() error{
	addr,err:= net.ResolveTCPAddr("tcp",Config.Config.ClusterConfig.LocalAddress)
	if err!=nil {
		return err
	}
	server := rpc.NewServer()
	this.serverListener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	this.rpcServer=server
	logger.Info(fmt.Sprintf("NodeComponent RPC server listening on: [ %s ]", addr.String()))
	go server.Accept(this.serverListener)
	return nil
}

func (this *NodeComponent)Register(rcvr interface{}) error {
	return this.rpcServer.Register(rcvr)
}

func (this *NodeComponent)clientCallback(event string,data ...interface{}) {
	switch event {
	case "close":
		nodeAddr := data[0].(string)
		this.rpcClient.Delete(nodeAddr)
		logger.Info(fmt.Sprintf("  disconnect to remote node: [ %s ]",nodeAddr))
	}
}

//获取节点客户端
func (this *NodeComponent) GetNodeClient(addr string) (*rpc.TcpClient,error){
	if v,ok:= this.rpcClient.Load(addr);ok {
		return v.(*rpc.TcpClient),nil
	}
	this.locker.Lock()
	defer this.locker.Unlock()
	if v,ok:= this.rpcClient.Load(addr);ok {
		return v.(*rpc.TcpClient),nil
	}
	client,err:= this.ConnectToNode(addr,this.clientCallback)
		if err!=nil {
		return nil,err
	}
	this.rpcClient.Store(addr,client)
	return client,nil
}

//查询并选择一个节点
func (this *NodeComponent)GetNode(role string,selectorType ...SelectorType) (*NodeID,error) {
	var nodeID *NodeID
	var  err error
	//优先查询位置服务器
	if this.islocationMode {
		nodeID, err = this.GetNodeFromLocation(role, selectorType...)
		if err == nil {
			return nodeID, nil
		}
	}
	//位置服务器不存在或不可用时在master上查询
	nodeID,err = this.GetNodeFromMaster(role,selectorType...)
	if err!=nil {
		return nil,err
	}
	return nodeID,nil
}

//查询节点组
func (this *NodeComponent)GetNodeGroup(role string) (*NodeIDGroup,error) {
	var nodeIDGroup *NodeIDGroup
	var err error
	//优先查询位置服务器
	if this.islocationMode {
		nodeIDGroup,err = this.GetNodeGroupFromLocation(role)
		if err==nil {
			return nodeIDGroup,nil
		}
	}

	//位置服务器不存在或不可用时在master上查询
	nodeIDGroup,err= this.GetNodeGroupFromMaster(role)
	if err!=nil {
		return nil,err
	}
	return nodeIDGroup,nil
}

//从位置服务器查询并选择一个节点
func (this *NodeComponent)GetNodeFromLocation(role string,selectorType ...SelectorType) (*NodeID,error) {
	var client *rpc.TcpClient
	var err error
	this.locker.Lock()
	if this.locationClients==nil{
		this.locker.Unlock()
		return nil, errors.New("location server not found")
	}
	client,err= this.locationClients.RandClient()
	this.locker.Unlock()
	if err!=nil {
		return nil,err
	}
	var reply *[]*InquiryReply
	args:=[]string{
		SELECTOR_TYPE_DEFAULT,Config.Config.ClusterConfig.AppName,role,
	}
	if len(selectorType)>0{
		args[0] = selectorType[0]
	}

	err = client.Call("LocationService.NodeInquiry",args,&reply)
	if err!=nil {
		this.locker.Lock()
		this.locationClients=nil
		this.locker.Unlock()
		return nil,err
	}
	if len(*reply)>0{
		g:=&NodeID{
			nodeComponent: this,
			Addr:          (*reply)[0].Node,
		}
		return g, nil
	}

	return nil,errors.New("no node of this role:"+role)
}

//从位置服务器查询
func (this *NodeComponent)GetNodeGroupFromLocation(role string) (*NodeIDGroup,error) {
	this.locker.Lock()
	client,err:= this.locationClients.RandClient()
	this.locker.Unlock()
	if err!=nil {
		return nil,err
	}
	var reply *[]*InquiryReply
	args:=[]string{
		SELECTOR_TYPE_GROUP,Config.Config.ClusterConfig.AppName,role,
	}
	err = client.Call("LocationService.NodeInquiry",args,&reply)
	if err!=nil {
		this.locker.Lock()
		this.locationClients=nil
		this.locker.Unlock()
		return nil,err
	}
	g:=&NodeIDGroup{
		nodeComponent:this,
		nodes:*reply,
	}
	return g, nil
}
//从master查询并选择一个节点
func (this *NodeComponent)GetNodeFromMaster(role string,selectorType ...SelectorType) (*NodeID,error) {
	if !this.IsOnline() {
		return nil,ErrNodeOffline
	}
	client,err:= this.GetNodeClient(Config.Config.ClusterConfig.MasterAddress)
	if err!=nil {
		return nil,err
	}
	var reply *[]*InquiryReply
	args:=[]string{
		SELECTOR_TYPE_DEFAULT,Config.Config.ClusterConfig.AppName,role,
	}
	if len(selectorType)>0{
		args[0] = selectorType[0]
	}
	err = client.Call("MasterService.NodeInquiry",args,&reply)
	if err!=nil {
		return nil,err
	}
	if len(*reply)>0{
		g:=&NodeID{
			nodeComponent: this,
			Addr:          (*reply)[0].Node,
		}
		return g, nil
	}

	return nil,errors.New("no node of this role:"+role)
}
//从master查询
func (this *NodeComponent)GetNodeGroupFromMaster(role string) (*NodeIDGroup,error) {
	if !this.IsOnline() {
		return nil,ErrNodeOffline
	}
	client,err:= this.GetNodeClient(Config.Config.ClusterConfig.MasterAddress)
	if err!=nil {
		return nil,err
	}
	var reply *[]*InquiryReply
	args:=[]string{
		SELECTOR_TYPE_GROUP,Config.Config.ClusterConfig.AppName,role,
	}
	err = client.Call("MasterService.NodeInquiry",args,&reply)
	if err!=nil {
		return nil,err
	}
	g:=&NodeIDGroup{
		nodeComponent:this,
		nodes:*reply,
	}
	return g, nil
}

//连接到某个节点
func (this *NodeComponent) ConnectToNode(addr string,callback func(event string,data ...interface{})) (*rpc.TcpClient, error) {
	client, err := rpc.NewTcpClient("tcp", addr,callback)
	if err != nil {
		return nil,err
	}
	count:=0
	for err!=nil {
		time.Sleep(time.Millisecond * 500)
		err = client.Reconnect()

		if err!=nil  {
			count++
			if count > 3{
				return nil,err
			}
		}
	}
	this.rpcClient.Store(addr,client)

	logger.Info(fmt.Sprintf("  connect to node: [ %s ] success",addr))
	return client,nil
}
