package RockGO

import (
	"github.com/zllangct/RockGO/actor"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/rpc"
	"runtime"
	"time"
)

var defaultNode *ServerNode

type ServerNode struct {
	Runtime        *Component.Runtime
	componentGroup *Component.ComponentGroups
	config 		   *Config.ConfigComponent
}

//新建一个服务节点
func NewServerNode() *ServerNode {
	return &ServerNode{
		componentGroup: &Component.ComponentGroups{},
		Runtime:        Component.NewRuntime(Component.Config{ThreadPoolSize: runtime.NumCPU()}),
	}
}
//获取默认节点
func DefaultNode() *ServerNode {
	if defaultNode != nil {
		return defaultNode
	}
	defaultNode = NewServerNode()
	return defaultNode
}
//开始服务
func (this *ServerNode) Serve() {
	//读取配置文件，初始化配置
	this.Runtime.Root().AddComponent(&Config.ConfigComponent{})
	rpc.CallTimeout = time.Millisecond * time.Duration(Config.Config.ClusterConfig.RpcCallTimeout)
	rpc.Timeout = time.Millisecond * time.Duration(Config.Config.ClusterConfig.RpcTimeout)
	rpc.HeartInterval = time.Millisecond * time.Duration(Config.Config.ClusterConfig.RpcHeartBeatInterval)
	rpc.DebugMode = Config.Config.CommonConfig.Debug
	//添加NodeComponent组件，使对象成为分布式节点
	this.Runtime.Root().AddComponent(&Cluster.NodeComponent{})
	//添加ActorProxy组件，组织节点间的通信
	this.Runtime.Root().AddComponent(&Actor.ActorProxyComponent{})
	//添加Actor组件，使该root节点成为可通讯的节点
	this.Runtime.Root().AddComponent(&Actor.ActorComponent{})
	//添加组件到待选组件列表，默认添加master,child组件
	this.AddComponentGroup("master",[]Component.IComponent{&Cluster.MasterComponent{}})
	this.AddComponentGroup("child",[]Component.IComponent{&Cluster.ChildComponent{}})
	//添加基础组件组,一般通过组建组的定义决定服务器节点的服务角色
	err:= this.componentGroup.AttachGroupsTo(this.GetConfig().ClusterConfig.Role, this.Runtime.Root())
	if err!=nil {
		logger.Fatal(err)
		panic(err)
	}
	//设置runtime工作线程
	this.Runtime.SetMaxThread(this.GetConfig().CommonConfig.RuntimeMaxWorker)
	go func() {
		var step float32=0
		for {
			step++
			this.Runtime.Update(step)
			time.Sleep(time.Millisecond * 500)
		}
	}()
}

//获取节点的Object根对象
func (this *ServerNode) GetRoot() *Component.Object {
	return this.Runtime.Root()
}
//添加一个组件组到组建组列表，不会立即添加到对象
func (this *ServerNode) AddComponentGroup(groupName string, group []Component.IComponent) {
	this.componentGroup.AddGroup(groupName, group)
}
//添加多个组件组到组建组列表，不会立即添加到对象
func (this *ServerNode) AddComponentGroups(groups map[string][]Component.IComponent) error {
	for groupName, group := range groups {
		this.componentGroup.AddGroup(groupName, group)
	}
	return nil
}
//获取配置
func (this *ServerNode) GetConfig() *Config.ConfigComponent {
	if this.config==nil {
		err := this.Runtime.Root().Find(&this.config)
		if err != nil {
			println(err.Error())
		}
	}
	return this.config
}
