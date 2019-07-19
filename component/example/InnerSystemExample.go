package main

import (
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/logger"
	"strconv"
	"sync"
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
	//新建一个运行时
	runtime:=Component.NewRuntime(Component.Config{
		ThreadPoolSize:20, //工作线程数量
	})

	//刷新帧
	go runtime.UpdateFrameByInterval(time.Millisecond*500)

	//新建一个对象实体
	o := Component.NewObject("o")
	err:=runtime.Root().AddObject(o)
	if err!=nil {
		logger.Error(err)
	}

	//给对象添加若干组件
	wg:=sync.WaitGroup{}
	wg.Add(1000)
	go func() {
		for i := 0; i<1000;i++  {
			_,_=o.AddNewObjectWithComponent(&TS{name:i},strconv.Itoa(i))
			wg.Done()
			time.Sleep(time.Millisecond*10)
		}
	}()

	//删除组件
	wg.Wait()
	go func() {
		for i := 0; i<1000;i++ {
			target,err:=o.FindObject(strconv.Itoa(i))
			if err!=nil {
				logger.Error(err)
				continue
			}
			err = target.Destroy()
			if err != nil {
				logger.Error(err)
			}
			time.Sleep(time.Millisecond*10)
		}
	}()

	time.Sleep(time.Second * 3)
}
