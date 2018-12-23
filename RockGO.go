package RockGO

import (
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/launcher"
	"runtime"
	"time"
)

var defaultNode *Server

type Server = launcher.ServerNode

//新建一个服务节点
func NewServerNode() *Server {
	active:=make(chan *Server)
	//构造运行时
	runtime:=Component.NewRuntime(Component.Config{ThreadPoolSize: runtime.NumCPU()})
	runtime.Root().AddComponent(&launcher.LauncherComponent{Active:active})
	runtime.UpdateFrameByInterval(time.Millisecond*100)
	return <-active
}

//获取默认节点
func DefaultServer() *Server {
	if defaultNode != nil {
		return defaultNode
	}
	defaultNode = NewServerNode()
	return defaultNode
}
