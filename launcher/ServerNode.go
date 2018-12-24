package launcher

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/actor"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/logger"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ServerNode struct {
	Runtime        *Component.Runtime
	componentGroup *Cluster.ComponentGroups
	Config         *Config.ConfigComponent
	Close          chan struct{}
}
func (this *ServerNode) Serve(){
	//添加NodeComponent组件，使对象成为分布式节点
	this.Root().AddComponent(&Cluster.NodeComponent{})

	//添加ActorProxy组件，组织节点间的通信
	if Config.Config.ClusterConfig.IsActorModel {
		this.Root().AddComponent(&Actor.ActorProxyComponent{})
	}

	//添加组件到待选组件列表，默认添加master,child组件
	this.AddComponentGroup("master",[]Component.IComponent{&Cluster.MasterComponent{}})
	this.AddComponentGroup("child",[]Component.IComponent{&Cluster.ChildComponent{}})
	if Config.Config.ClusterConfig.IsLocationMode {
		this.AddComponentGroup("location",[]Component.IComponent{&Cluster.LocationComponent{}})
	}

	//添加基础组件组,一般通过组建组的定义决定服务器节点的服务角色
	err:= this.componentGroup.AttachGroupsTo(Config.Config.ClusterConfig.Role, this.Root())
	if err!=nil {
		logger.Fatal(err)
		panic(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c,syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	//go func() {
	//	time.Sleep(time.Second*2)
	//	this.Close<- struct {}{}
	//}()
	for  {
		select {
		case <-c:
		case <-this.Close:
		}
		logger.Info("====== Start to close this server, do some cleaning now ...... ======")
		//do something else
		err=this.Root().Destroy()
		if err!=nil{
			logger.Error(err)
		}
		time.Sleep(time.Second*2)
		//close success
		logger.Info("====== Server is closed ======")
		return
	}
}
//获取节点的Object根对象
func (this *ServerNode) Root() *Component.Object {
	if this.Config==nil {
		panic(ErrServerNotInit)
	}
	return this.Runtime.Root()
}

//覆盖节点信息
func (this *ServerNode) OverrideNodeDefine(nodeConfName string)error{
	if this.Config==nil {
		panic(ErrServerNotInit)
	}
	if s,ok:= this.Config.ClusterConfig.NodeDefine[nodeConfName];ok{
		this.Config.ClusterConfig.LocalAddress = s.LocalAddress
		this.Config.ClusterConfig.Role =s.Role
	}else{
		return errors.New(fmt.Sprintf("this config name [ %s ] not defined",nodeConfName))
	}
	return nil
}

//添加一个组件组到组建组列表，不会立即添加到对象
func (this *ServerNode) AddComponentGroup(groupName string, group []Component.IComponent) {
	if this.Config==nil {
		panic(ErrServerNotInit)
	}
	if Config.Config.ClusterConfig.IsActorModel {
		group= append(group, &Actor.ActorComponent{})
	}
	this.componentGroup.AddGroup(groupName, group)
}

//添加多个组件组到组建组列表，不会立即添加到对象
func (this *ServerNode) AddComponentGroups(groups map[string][]Component.IComponent) error {
	if this.Config==nil {
		panic(ErrServerNotInit)
	}
	for groupName, group := range groups {
		if Config.Config.ClusterConfig.IsActorModel {
			group= append(group,&Actor.ActorComponent{})
		}
		this.componentGroup.AddGroup(groupName, group)
	}
	return nil
}
