package Cluster

import (
	"github.com/zllangct/RockGO/actor"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"runtime"
)

var defaultNode *ServerNode

type ServerNode struct {
	componentGroup *Component.ComponentGroups
	Runtime        *Component.Runtime
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

func (this *ServerNode) Serve() {
	//读取配置文件，初始化配置
	this.Runtime.Root().AddComponent(&Config.ConfigComponent{})
	config := this.GetConfig()
	//设置runtime工作线程
	this.Runtime.SetMaxThread(config.CommonConfig.RuntimeMaxWorker)
	//添加ActorProxy组件，组织节点间的通信
	this.Runtime.Root().AddComponent(&Actor.ActorProxyComponent{})
	//添加Actor组件，使该节点成为可通讯的节点
	this.Runtime.Root().AddComponent(&Actor.ActorComponent{})
	//添加组件到待选组件列表，默认添加master组件
	this.AddComponentGroup("master",[]Component.IComponent{&MasterComponent{}})
	//添加基础组件组,一般通过组建组的定义决定服务器节点的服务角色
	this.componentGroup.AttachGroupTo(config.CusterConfig.Group, this.Runtime.Root())

}

//获取节点的Object根对象
func (this *ServerNode) GetRoot() *Component.Object {
	return this.Runtime.Root()
}
//添加一个组件组到组建组列表，不会立即添加到对象
func (this *ServerNode) AddComponentGroup(groupName string, group Component.ComponentGroup) error {
	return this.componentGroup.AddGroup(groupName, group)
}
//添加多个组件组到组建组列表，不会立即添加到对象
func (this *ServerNode) AddComponentGroups(groups map[string]Component.ComponentGroup) error {
	for groupName, group := range groups {
		err := this.componentGroup.AddGroup(groupName, group)
		if err != nil {
			return err
		}
	}
	return nil
}

//获取配置
func (this *ServerNode) GetConfig() *Config.ConfigComponent {
	var config *Config.ConfigComponent
	err := this.Runtime.Root().Find(&config)
	if err != nil {
		println(err.Error())
	}
	return config
}
