package main

import (
	"github.com/zllangct/RockGO/ecs"
	"github.com/zllangct/RockGO/logger"
	"time"
)

type TS struct {
	ecs.ComponentBase
	name int
}

func (this *TS) Start(ctx *ecs.Context) {
	println("start: ", this.name)
}

func (this *TS) Update(ctx *ecs.Context) {
	println("update: ", this.name)
}

func (this *TS) Awake(ctx *ecs.Context) {
	println("awake: ", this.name)
}

func (this *TS) Destroy(ctx *ecs.Context) {
	println("destroy: ", this.name)
}

func main() {
	runtime := ecs.NewRuntime(ecs.Config{
		ThreadPoolSize: 20,
	})

	go runtime.UpdateFrameByInterval(time.Millisecond * 500)

	o := ecs.NewObject("o")
	_ = runtime.Root().AddObject(o)

	o1 := ecs.NewObject("o1")
	o2 := ecs.NewObject()

	_ = o.AddObject(o1)
	_ = o.AddObject(o2)

	//对象树
	println(runtime.Root().Debug(2))

	//查找对象
	oo, err := o.FindObject("o1")
	if err != nil {
		logger.Error(err)
	}
	logger.Info(oo.Name())

	//添加组件
	cpt := &TS{name: 1234}
	o1.AddComponent(cpt)

	//查找组件
	var temp *TS
	err = o1.Find(&temp)
	if err != nil {
		logger.Error(err)
	}
	println(temp.name)

	time.Sleep(time.Second * 2)

	//移除组件
	o1.RemoveComponent(temp)

	time.Sleep(time.Second * 2)

	err = o.RemoveObject(o2)
	if err != nil {
		logger.Error()
	}
	println(runtime.Root().Debug(2))
}
