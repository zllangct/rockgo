package Cluster

import (
	"github.com/zllangct/RockGO/Component"
	"github.com/zllangct/RockGO/Config"
	"runtime"
)

type ServerNode struct {
	Runtime *Component.Runtime
}

//新建一个服务节点
func NewServerNode() *ServerNode  {
	return &ServerNode{
		Runtime:Component.NewRuntime(Component.Config{ThreadPoolSize:runtime.NumCPU()}),
	}
}

func (this *ServerNode)Serve()  {
	//读取配置文件，初始化配置
	this.Runtime.Root().AddComponent(&Config.ConfigComponent{})
	//设置runtime工作线程
	this.Runtime.SetMaxThread(this.GetConfig().CommonConfig.RuntimeMaxWorker)
}

//获取配置
func (this *ServerNode)GetConfig() (*Config.ConfigComponent){
	var config *Config.ConfigComponent
	err:=this.Runtime.Root().Find(&config)
	if err!=nil{
		println(err.Error())
	}
	return config
}