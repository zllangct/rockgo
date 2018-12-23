package RockGO

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/actor"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/rpc"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var Server *ServerNode
var defaultNode *ServerNode

var ErrServerNotInit =errors.New("server is not initialize")

type ServerNode struct {
	Runtime        *Component.Runtime
	componentGroup *Cluster.ComponentGroups
	Config         *Config.ConfigComponent
	initialized         bool
	Close          chan struct{}
}

//新建一个服务节点
func NewServerNode() *ServerNode {
	s:= &ServerNode{
		Close:make(chan struct{}),
		componentGroup: &Cluster.ComponentGroups{},
		Runtime:        Component.NewRuntime(Component.Config{ThreadPoolSize: runtime.NumCPU()}),
	}
	s.Init()
	return s
}

//获取默认节点
func DefaultNode() *ServerNode {
	if defaultNode != nil {
		return defaultNode
	}
	defaultNode = NewServerNode()
	return defaultNode
}

func (this *ServerNode)Init()  {
	//读取配置文件，初始化配置
	this.Runtime.Root().AddComponent(&Config.ConfigComponent{})
	this.Config=Config.Config
	//设置runtime工作线程
	this.Runtime.SetMaxThread(Config.Config.CommonConfig.RuntimeMaxWorker)
	//rpc
	rpc.CallTimeout = time.Millisecond * time.Duration(Config.Config.ClusterConfig.RpcCallTimeout)
	rpc.Timeout = time.Millisecond * time.Duration(Config.Config.ClusterConfig.RpcTimeout)
	rpc.HeartInterval = time.Millisecond * time.Duration(Config.Config.ClusterConfig.RpcHeartBeatInterval)
	rpc.DebugMode = Config.Config.CommonConfig.Debug
	//log
	switch Config.Config.CommonConfig.LogMode {
	case logger.DAILY:
		logger.SetRollingDaily(Config.Config.CommonConfig.LogPath, Config.Config.ClusterConfig.AppName+".log")
	case logger.ROLLFILE:
		logger.SetRollingFile(Config.Config.CommonConfig.LogPath, Config.Config.ClusterConfig.AppName+".log",
			1000, Config.Config.CommonConfig.LogFileMax, Config.Config.CommonConfig.LogFileUnit)
	}
	logger.SetLevel(Config.Config.CommonConfig.LogLevel)
	this.initialized =true
}


//开始服务
func (this *ServerNode) Serve(){
	if !this.initialized {
		panic(ErrServerNotInit)
	}
	//添加NodeComponent组件，使对象成为分布式节点
	this.Runtime.Root().AddComponent(&Cluster.NodeComponent{})
	//添加ActorProxy组件，组织节点间的通信
	if Config.Config.ClusterConfig.IsActorModel {
		this.Runtime.Root().AddComponent(&Actor.ActorProxyComponent{})
	}
	//添加组件到待选组件列表，默认添加master,child组件
	this.AddComponentGroup("master",[]Component.IComponent{&Cluster.MasterComponent{}})
	this.AddComponentGroup("child",[]Component.IComponent{&Cluster.ChildComponent{}})
	if Config.Config.ClusterConfig.IsLocationMode {
		this.AddComponentGroup("location",[]Component.IComponent{&Cluster.LocationComponent{}})
	}

	//添加基础组件组,一般通过组建组的定义决定服务器节点的服务角色
	err:= this.componentGroup.AttachGroupsTo(Config.Config.ClusterConfig.Role, this.Runtime.Root())
	if err!=nil {
		logger.Fatal(err)
		panic(err)
	}

	go this.Runtime.UpdateFrameByInterval(time.Millisecond * 100)

	c := make(chan os.Signal)
	signal.Notify(c,syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	for  {
		select {
		case <-c:
			logger.Info("====== Start to close this server, do some cleaning now ...... ======")
		case <-this.Close:
		}
		//do something else
		err=this.Runtime.Root().Destroy()
		if err!=nil{
			logger.Error(err)
		}
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
//获取配置
func (this *ServerNode) GetConfig() *Config.ConfigComponent {
	if this.Config==nil {
		panic(ErrServerNotInit)
	}
	if this.Config ==nil {
		err := this.Runtime.Root().Find(&this.Config)
		if err != nil {
			logger.Error(err)
		}
	}
	return this.Config
}
