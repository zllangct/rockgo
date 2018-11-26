package Actor

import (
	"github.com/zllangct/RockGO/Component"
	"sync"
)


type ActorAddressServer struct {
	Component.Base
	ActorApp sync.Map
	ActorRule sync.Map
}


func (this *ActorAddressServer)Awake()  {
	
}

func (this *ActorAddressServer)Start(ctx *Component.Context)  {

}

func (this *ActorAddressServer)Registor(actorID ActorID)  {
	addressArr:=actorID.ToArray()
	//ActorAddress : all apps

	//rule
	m_rule:=&sync.Map{}
	if m,ok:=this.ActorAddress.LoadOrStore(addressArr[ADRESS_APPID],m_rule);!ok {
		m_rule=m.(*sync.Map)
	}
	//node
	m_node:=&sync.Map{}
	if m,ok:=this.ActorAddress.LoadOrStore(addressArr[ADRESS_APPID],m_node);!ok {
		m_node=m.(*sync.Map)
	}
}

func (this *ActorAddressServer)Unregistor(actorID ActorID)  {

}