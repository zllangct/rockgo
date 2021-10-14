package main

import (
	"flag"
	"fmt"
	"github.com/zllangct/rockgo"
	"github.com/zllangct/rockgo/ecs"
	"github.com/zllangct/rockgo/gate"
	"github.com/zllangct/rockgo/logger"
	"net/http"
	_ "net/http/pprof"
)

/*
	单服、多角色实例，本实例包括 网关角色和登录角色，诸如，大厅角色、房间角色同理配置。
	使用component，传统方式搭建。
*/
var Server *RockGO.Server

func main() {
	var nodeConfName string
	flag.StringVar(&nodeConfName, "node", "", "node info")
	flag.Parse()

	/*  pprof */
	go func() {
		logger.Info(http.ListenAndServe("localhost:7077", nil))
	}()

	Server = RockGO.DefaultServer()

	//重选节点
	if nodeConfName != "" {
		Server.OverrideNodeDefine(nodeConfName)
		logger.Info(fmt.Sprintf("Override node info:[ %s ]", nodeConfName))
	}

	g := &gate.DefaultGateComponent{}

	//非默认网关不必要这样初始化，在组件内初始化即可，此处是为了对默认网关注入
	g.AddNetAPI(NewTestApi())

	Server.AddComponentGroup("gate", []ecs.IComponent{g})
	Server.AddComponentGroup("room", []ecs.IComponent{&RoomManagerComponent{}})
	Server.Serve()
}
