package main

import (
	"github.com/zllangct/RockGO"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/gate"
)

func main()  {
	RockGO.Server = RockGO.DefaultNode()
	RockGO.Server.AddComponentGroup("gate",[]Component.IComponent{&gate.DefaultGateComponent{
		NetAPI:NewTestApi(),
	}})
	RockGO.Server.AddComponentGroup("login",[]Component.IComponent{&LoginComponent{}})
	RockGO.Server.Serve()
}
