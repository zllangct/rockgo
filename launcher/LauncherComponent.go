package launcher

import (
	"errors"
	"github.com/zllangct/RockGO/actor"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/rpc"
	"time"
)
var ErrServerNotInit =errors.New("server is not initialize")

/* 服务端启动组件 */
type LauncherComponent struct {
	Component.Base
	server *ServerNode
	Active chan *ServerNode
}

func (this *LauncherComponent) IsUnique() int {
	return Component.UNIQUE_TYPE_GLOBAL
}

func (this *LauncherComponent) Awake(context *Component.Context) {
	//新建server
	this.server=&ServerNode{
		Close:make(chan struct{}),
		componentGroup:&Cluster.ComponentGroups{},
		Runtime:this.Runtime(),
	}
	//读取配置文件，初始化配置
	this.Root().AddComponent(&Config.ConfigComponent{})

	//缓存配置文件
	this.server.Config=Config.Config

	//设置runtime工作线程
	this.Runtime().SetMaxThread(Config.Config.CommonConfig.RuntimeMaxWorker)

	//rpc设置
	rpc.CallTimeout = time.Millisecond * time.Duration(Config.Config.ClusterConfig.RpcCallTimeout)
	rpc.Timeout = time.Millisecond * time.Duration(Config.Config.ClusterConfig.RpcTimeout)
	rpc.HeartInterval = time.Millisecond * time.Duration(Config.Config.ClusterConfig.RpcHeartBeatInterval)
	rpc.DebugMode = Config.Config.CommonConfig.Debug

	//log设置
	switch Config.Config.CommonConfig.LogMode {
	case logger.DAILY:
		logger.SetRollingDaily(Config.Config.CommonConfig.LogPath, Config.Config.ClusterConfig.AppName+".log")
	case logger.ROLLFILE:
		logger.SetRollingFile(Config.Config.CommonConfig.LogPath, Config.Config.ClusterConfig.AppName+".log",
			1000, Config.Config.CommonConfig.LogFileMax, Config.Config.CommonConfig.LogFileUnit)
	}
	logger.SetLevel(Config.Config.CommonConfig.LogLevel)

	//添加NodeComponent组件，使对象成为分布式节点
	this.Root().AddComponent(&Cluster.NodeComponent{})

	//添加ActorProxy组件，组织节点间的通信
	if Config.Config.ClusterConfig.IsActorModel {
		this.Root().AddComponent(&Actor.ActorProxyComponent{})
	}

	this.Active<-this.server
}





