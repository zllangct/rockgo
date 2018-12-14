package main

import (
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
	go func() {
		logger.Info(http.ListenAndServe("localhost:7070", nil))
	}()
	RockGO.Server = RockGO.DefaultNode()
	RockGO.Server.AddComponentGroup("gate",[]Component.IComponent{&gate.DefaultGateComponent{
		NetAPI:NewTestApi(),
	}})
	RockGO.Server.AddComponentGroup("room",[]Component.IComponent{&RoomManagerComponent{}})
	RockGO.Server.Serve()
}
