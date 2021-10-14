package main

import (
	"github.com/zllangct/rockgo/ecs"
	"github.com/zllangct/rockgo/logger"
	"strconv"
	"sync"
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
	//新建一个运行时
	runtime := ecs.NewRuntime(ecs.Config{
		ThreadPoolSize: 20, //工作线程数量
	})

	//刷新帧
	go runtime.UpdateFrameByInterval(time.Millisecond * 500)

	//新建一个对象实体
	o := ecs.NewObject("o")
	err := runtime.Root().AddObject(o)
	if err != nil {
		logger.Error(err)
	}

	//给对象添加若干组件
	wg := sync.WaitGroup{}
	wg.Add(1000)
	go func() {
		for i := 0; i < 1000; i++ {
			_, _ = o.AddNewObjectWithComponent(&TS{name: i}, strconv.Itoa(i))
			wg.Done()
			time.Sleep(time.Millisecond * 10)
		}
	}()

	//删除组件
	wg.Wait()
	go func() {
		for i := 0; i < 1000; i++ {
			target, err := o.FindObject(strconv.Itoa(i))
			if err != nil {
				logger.Error(err)
				continue
			}
			err = target.Destroy()
			if err != nil {
				logger.Error(err)
			}
			time.Sleep(time.Millisecond * 10)
		}
	}()

	time.Sleep(time.Second * 3)
}
