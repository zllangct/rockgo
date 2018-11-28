package Actor

import (
	"github.com/zllangct/RockGO/component"
	"sync"
)


type ActorAddressServerComponent struct {
	Component.Base
	ActorApp sync.Map
	ActorRule sync.Map
}


func (this *ActorAddressServerComponent)Awake()  {
	
}

func (this *ActorAddressServerComponent)Start(ctx *Component.Context)  {

}

func (this *ActorAddressServerComponent)Registor(actorID ActorID)  {
	//addressArr:=actorID.ToArray()
	//ActorAddress : all apps

	////rule
	//m_rule:=&sync.Map{}
	//if m,ok:=this.ActorAddress.LoadOrStore(addressArr[ADRESS_APPID],m_rule);!ok {
	//	m_rule=m.(*sync.Map)
	//}
	////node
	//m_node:=&sync.Map{}
	//if m,ok:=this.ActorAddress.LoadOrStore(addressArr[ADRESS_APPID],m_node);!ok {
	//	m_node=m.(*sync.Map)
	//}
}

func (this *ActorAddressServerComponent)Unregistor(actorID ActorID)  {

}