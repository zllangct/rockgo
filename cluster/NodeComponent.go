package Cluster

import (
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
	isOnline       	bool
	rpcClient      	sync.Map 				//RPC客户端集合
	rpcServer      	*rpc.Server				//本节点RPC Server
	config 			*Config.ConfigComponent
}

func (this *NodeComponent) GetRequire() (map[*Component.Object][]reflect.Type) {
	requires:=make(map[*Component.Object][]reflect.Type)
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
	}
	return requires
}

func(this *NodeComponent)Awake(){
	err:= this.Parent.Root().Find(&this.config)
	if err != nil {
		logger.Error("get config component failed")
		panic(err)
		return
	}
	//开始本节点RPC服务
	this.StartRpcServer()
}

//RPC服务
func (this *NodeComponent) StartRpcServer() error{
	var config *Config.ConfigComponent
	err := this.Parent.Root().Find(&config)
	if err != nil {
		return err
	}
	addr,err:= net.ResolveTCPAddr("tcp",config.ClusterConfig.LocalAddress)
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

//获取节点客户端
func (this *NodeComponent) GetNodeClient(addr string) (*rpc.TcpClient,error){

	return nil,nil
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
