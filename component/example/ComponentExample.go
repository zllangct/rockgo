package main

import (
	"github.com/zllangct/RockGO"
	"github.com/zllangct/RockGO/logger"
	"time"
)

type TS struct {
	Component.Base
	name int
}
func (this *TS)Start(ctx *Component.Context)  {
	println("start: ",this.name)
}

func (this *TS)Update(ctx *Component.Context)  {
	println("update: ",this.name)
}

func (this *TS)Awake(ctx *Component.Context)  {
	println("awake: ",this.name)
}

func (this *TS)Destroy(ctx *Component.Context)  {
	println("destroy: ",this.name)
}

func main()  {
	runtime:=Component.NewRuntime(Component.Config{
		ThreadPoolSize:20,
	})

	go runtime.UpdateFrameByInterval(time.Millisecond*500)

	o := Component.NewObject("o")
	_=runtime.Root().AddObject(o)

	o1:=Component.NewObject("o1")
	o2:=Component.NewObject()

	_=o.AddObject(o1)
	_=o.AddObject(o2)

	//对象树
	println(runtime.Root().Debug(2))

	//查找对象
	oo,err:=o.FindObject("o1")
	if err!=nil {
		logger.Error(err)
	}
	logger.Info(oo.Name())

	//添加组件
	cpt:=&TS{name:1234}
	o1.AddComponent(cpt)

	//查找组件
	var temp *TS
	err=o1.Find(&temp)
	if err!=nil {
		logger.Error(err)
	}
	println(temp.name)

	time.Sleep(time.Second *2)

	//移除组件
	o1.RemoveComponent(temp)

	time.Sleep(time.Second *2)

	err=o.RemoveObject(o2)
	if err!=nil {
		logger.Error()
	}
	println(runtime.Root().Debug(2))
}