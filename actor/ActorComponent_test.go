package Actor_test

import (
	"github.com/zllangct/RockGO/actor"
	"github.com/zllangct/RockGO/component"
	"testing"
)

func TestActorComponents(T *testing.T) {
	runtime := Component.NewRuntime(Component.Config{
		ThreadPoolSize: 10})

	o1:=Component.NewObject("actor1")
	o1.AddComponent(&Actor.ActorComponent{})

	runtime.Root().AddObject(o1)

	for i := 0; i<1;i++  {
		runtime.Update(float32(i))
	}
	var c *Actor.ActorComponent
	err:=o1.Find(c)
	if err!=nil{
		panic(err)
	}
	//message:=&actor.ActorMessageInfo{
	//	Message:
	//}
	//c.Tell()
}
