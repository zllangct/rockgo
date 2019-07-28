package launcher

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/actor"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/ecs"
	"github.com/zllangct/RockGO/config"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/rpc"
	"github.com/zllangct/RockGO/timer"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var ErrServerNotInit = errors.New("server is not initialize")

/* 服务端启动组件 */
type LauncherComponent struct {
	ecs.ComponentBase
	componentGroup *Cluster.ComponentGroups
	Config         *config.ConfigComponent
	Close          chan struct{}
}

func (this *LauncherComponent) IsUnique() int {
	return ecs.UNIQUE_TYPE_GLOBAL
}

func (this *LauncherComponent) Initialize() error {
	//新建server
	this.Close = make(chan struct{})
	this.componentGroup = &Cluster.ComponentGroups{}

	//读取配置文件，初始化配置
	this.Root().AddComponent(&config.ConfigComponent{})

	//缓存配置文件
	this.Config = config.Config

	//设置runtime工作线程
	this.Runtime().SetMaxThread(config.Config.CommonConfig.RuntimeMaxWorker)

	//rpc设置
	rpc.CallTimeout = time.Millisecond * time.Duration(config.Config.ClusterConfig.RpcCallTimeout)
	rpc.Timeout = time.Millisecond * time.Duration(config.Config.ClusterConfig.RpcTimeout)
	rpc.HeartInterval = time.Millisecond * time.Duration(config.Config.ClusterConfig.RpcHeartBeatInterval)
	rpc.DebugMode = config.Config.CommonConfig.Debug

	//log设置
	switch config.Config.CommonConfig.LogMode {
	case logger.DAILY:
		logger.SetRollingDaily(config.Config.CommonConfig.LogPath, config.Config.ClusterConfig.AppName+".log")
	case logger.ROLLFILE:
		logger.SetRollingFile(config.Config.CommonConfig.LogPath, config.Config.ClusterConfig.AppName+".log",
			1000, config.Config.CommonConfig.LogFileMax, config.Config.CommonConfig.LogFileUnit)
	}
	logger.SetLevel(config.Config.CommonConfig.LogLevel)
	return nil
}

func (this *LauncherComponent) Serve() {
	//添加NodeComponent组件，使对象成为分布式节点
	this.Root().AddComponent(&Cluster.NodeComponent{})

	//添加ActorProxy组件，组织节点间的通信
	this.Root().AddComponent(&Actor.ActorProxyComponent{})

	//添加组件到待选组件列表，默认添加master,child组件
	this.AddComponentGroup("master", []ecs.IComponent{&Cluster.MasterComponent{}})
	this.AddComponentGroup("child", []ecs.IComponent{&Cluster.ChildComponent{}})
	if config.Config.ClusterConfig.IsLocationMode && config.Config.ClusterConfig.Role[0] != "single" {
		this.AddComponentGroup("location", []ecs.IComponent{&Cluster.LocationComponent{}})
	}

	//处理single模式
	if len(config.Config.ClusterConfig.Role) == 0 || config.Config.ClusterConfig.Role[0] == "single" {
		config.Config.ClusterConfig.Role = this.componentGroup.AllGroupsName()
	}

	//添加基础组件组,一般通过组建组的定义决定服务器节点的服务角色
	err := this.componentGroup.AttachGroupsTo(config.Config.ClusterConfig.Role, this.Root())
	if err != nil {
		logger.Fatal(err)
		panic(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	/* 清理测试代码,ide关闭信号无法命中断点 */
	//go func() {
	//	time.Sleep(time.Second*2)
	//	this.Close<- struct {}{}
	//}()

	//等待服务器关闭，并执行停机处理
	select {
	case <-c:
	case <-this.Close:
	}

	logger.Info("====== Start to close this server, do some cleaning now ...... ======")
	//do something else
	err = this.Root().Destroy()
	if err != nil {
		logger.Error(err)
	}
	<-timer.After(time.Second)
	logger.Info("====== Server is closed ======")
}

//覆盖节点信息
func (this *LauncherComponent) OverrideNodeDefine(nodeConfName string) {
	if this.Config == nil {
		panic(ErrServerNotInit)
	}
	if s, ok := this.Config.ClusterConfig.NodeDefine[nodeConfName]; ok {
		this.Config.ClusterConfig.LocalAddress = s.LocalAddress
		this.Config.ClusterConfig.Role = s.Role
	} else {
		panic(errors.New(fmt.Sprintf("this config name [ %s ] not defined", nodeConfName)))
	}
}

//覆盖节点端口
func (this *LauncherComponent) OverrideNodePort(port string) {
	if this.Config == nil {
		panic(ErrServerNotInit)
	}
	ip := strings.Split(this.Config.ClusterConfig.LocalAddress, ":")[0]
	this.Config.ClusterConfig.LocalAddress = fmt.Sprintf("%s:%s", ip, port)
}

//覆盖节点角色
func (this *LauncherComponent) OverrideNodeRoles(roles []string) {
	if this.Config == nil {
		panic(ErrServerNotInit)
	}
	this.Config.ClusterConfig.Role = roles
}

//添加一个组件组到组建组列表，不会立即添加到对象
func (this *LauncherComponent) AddComponentGroup(groupName string, group []ecs.IComponent) {
	if this.Config == nil {
		panic(ErrServerNotInit)
	}
	this.componentGroup.AddGroup(groupName, group)
}

//添加多个组件组到组建组列表，不会立即添加到对象
func (this *LauncherComponent) AddComponentGroups(groups map[string][]ecs.IComponent) error {
	if this.Config == nil {
		panic(ErrServerNotInit)
	}
	for groupName, group := range groups {
		this.componentGroup.AddGroup(groupName, group)
	}
	return nil
}
