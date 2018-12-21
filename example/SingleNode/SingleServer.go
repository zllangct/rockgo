package main

import (
	"flag"
	"fmt"
	"github.com/zllangct/RockGO"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/gate"
	"github.com/zllangct/RockGO/logger"
)
/*
	单服、多角色实例，本实例包括 网关角色和登录角色，诸如，大厅角色、房间角色同理配置。
	使用component，传统方式搭建。
*/

func main()  {
	var nodeConfName string
	flag.StringVar(&nodeConfName,"node", "", "node info")
	flag.Parse()

	RockGO.Server = RockGO.DefaultNode()

	//重选节点
	if nodeConfName!=""{
		err:=RockGO.Server.OverrideNodeDefine(nodeConfName)
		if err!=nil {
			logger.Fatal(err)
		}
		logger.Info(fmt.Sprintf("Override node info:[ %s ]",nodeConfName))
	}

	g:=&gate.DefaultGateComponent{}
	g.NetAPI=NewTestApi(g.Parent)  //非默认网关不必要这样初始化，在组件内初始化即可，此处是为了对默认网关注入
	RockGO.Server.AddComponentGroup("gate",[]Component.IComponent{g})
	RockGO.Server.AddComponentGroup("login",[]Component.IComponent{&LoginComponent{}})
	RockGO.Server.Serve()
}
