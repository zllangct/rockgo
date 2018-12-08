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


type NodeComponent struct {
	Component.Base
	locker sync.RWMutex
	AppName         string
	isOnline       	bool
	rpcClient      	sync.Map 				//RPC客户端集合
	rpcServer      	*rpc.Server				//本节点RPC Server
	locationClients *NodeIDGroup				//位置服务器集合
}

func (this *NodeComponent) GetRequire() (map[*Component.Object][]reflect.Type) {
	requires:=make(map[*Component.Object][]reflect.Type)
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
	}
	return requires
}

func(this *NodeComponent)Awake(){
	this.AppName = Config.Config.ClusterConfig.AppName
	//开始本节点RPC服务
	this.StartRpcServer()
	//查询位置服务器
	go this.GetLocationServer()
}

func (this *NodeComponent)IsOnline() bool {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return this.isOnline
}

//获取位置服务器
func (this *NodeComponent)GetLocationServer()  {
	for {
		this.locker.RLock()
		online:=this.isOnline
		this.locker.RUnlock()
		if !online {
			time.Sleep(time.Second)
			continue
		}
		g,err:= this.GetNodeGroupFromMaster("location")
		if err!=nil {
			logger.Debug(err)
			time.Sleep(time.Second)
			continue
		}
		this.locker.Lock()
		this.locationClients = g
		this.locker.Unlock()
	}
}

//RPC服务
func (this *NodeComponent) StartRpcServer() error{
	addr,err:= net.ResolveTCPAddr("tcp",Config.Config.ClusterConfig.LocalAddress)
	if err!=nil {
		return err
	}
	server := rpc.NewServer()
	var l net.Listener
	l, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	this.rpcServer=server
	logger.Info("Test RPC server listening on", addr.String())
	go server.Accept(l)
	return nil
}

func (this *NodeComponent)clientCallback(event string,data ...interface{}) {
	switch event {
	case "close":
		nodeAddr := data[0].(string)
		this.rpcClient.Delete(nodeAddr)
	}
}

//获取节点客户端
func (this *NodeComponent) GetNodeClient(addr string) (*rpc.TcpClient,error){
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
	//优先查询位置服务器
	nodeID,err:= this.GetNodeFromLocation(role,selectorType...)
	if err==nil {
		return nodeID,nil
	}
	//位置服务器不存在或不可用时在master上查询
	nodeID,err= this.GetNodeFromMaster(role,selectorType...)
	if err!=nil {
		return nil,err
	}
	return nodeID,nil
}

//从位置服务器查询并选择一个节点
func (this *NodeComponent)GetNodeFromLocation(role string,selectorType ...SelectorType) (*NodeID,error) {
	this.locker.Lock()
	client,err:= this.locationClients.RandClient()
	this.locker.Unlock()
	if err!=nil {
		return nil,err
	}
	var reply *[]*InquiryReply
	args:=fmt.Sprintf("%s:%s",this.AppName,"location")
	if len(selectorType)>0{
		args=fmt.Sprintf("%s:%s",args,string(selectorType[0]))
	}
	err = client.Call("LocationService.NodeInquiry",args,&reply)
	if err!=nil {
		return nil,err
	}
	if len(*reply)>0{
		g:=&NodeID{
			nodeComponent:this,
			addr:(*reply)[0].Node,
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
	err = client.Call("LocationService.NodeInquiry",fmt.Sprintf("%s:%s",this.AppName,"location"),&reply)
	if err!=nil {
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
	client,err:= this.GetNodeClient(Config.Config.ClusterConfig.MasterAddress)
	if err!=nil {
		return nil,err
	}
	var reply *[]*InquiryReply
	args:=fmt.Sprintf("%s:%s",this.AppName,"location")
	if len(selectorType)>0{
		args=fmt.Sprintf("%s:%s",args,string(selectorType[0]))
	}
	err = client.Call("MasterService.NodeInquiryDetail",args,&reply)
	if err!=nil {
		return nil,err
	}
	if len(*reply)>0{
		g:=&NodeID{
			nodeComponent:this,
			addr:(*reply)[0].Node,
		}
		return g, nil
	}

	return nil,errors.New("no node of this role:"+role)
}
//从master查询
func (this *NodeComponent)GetNodeGroupFromMaster(role string) (*NodeIDGroup,error) {
	client,err:= this.GetNodeClient(Config.Config.ClusterConfig.MasterAddress)
	if err!=nil {
		return nil,err
	}
	var reply *[]*InquiryReply
	err = client.Call("MasterService.NodeInquiryDetail",fmt.Sprintf("%s:%s",this.AppName,"location"),&reply)
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

	logger.Info(time.Now().Format("2006-01-02T 15:04:05")+fmt.Sprintf("  connect to node: [ %s ] success",addr))
	return client,nil
}
