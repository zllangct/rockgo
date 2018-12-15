package main

import (
	"flag"
	"fmt"
	"github.com/zllangct/RockGO"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/gate"
	"github.com/zllangct/RockGO/logger"
	"net/http"
	_ "net/http/pprof"
)
/*
	单服、多角色实例，本实例包括 网关角色和登录角色，诸如，大厅角色、房间角色同理配置。
	使用component，传统方式搭建。
*/

func main()  {
	var nodeConfName string
	flag.StringVar(&nodeConfName,"node", "", "node info")
	flag.Parse()

	go func() {
		logger.Info(http.ListenAndServe("localhost:7070", nil))
	}()
	RockGO.Server = RockGO.DefaultNode()

	//重选节点
	if nodeConfName!=""{
		err:=RockGO.Server.OverrideNodeDefine(nodeConfName)
		if err!=nil {
			logger.Fatal(err)
		}
		logger.Info(fmt.Sprintf("Override node info:[ %s ]",nodeConfName))
	}

	RockGO.Server.AddComponentGroup("gate",[]Component.IComponent{&gate.DefaultGateComponent{
		NetAPI:NewTestApi(),
	}})
	RockGO.Server.AddComponentGroup("room",[]Component.IComponent{&RoomManagerComponent{}})
	RockGO.Server.Serve()
}
