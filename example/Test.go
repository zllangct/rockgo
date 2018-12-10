package main

import (
	"github.com/zllangct/RockGO/3rd/threadpool"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/logger"
	"net/http"
	_ "net/http/pprof"
	"reflect"
	"strconv"
	"sync"
	"time"
)

type Hello struct {
}

func (this *Hello) Hello(str string) {

}
func (this *Hello) Type() reflect.Type {
	return reflect.TypeOf(this)
}

func (this *Hello) Start(context *Component.Context) {
	this.Hello("my name is Rock.")

}

func (this *Hello) Update(context *Component.Context) {
	this.Hello(strconv.Itoa(1))
}

func main() {
	go func() {
		logger.Info(http.ListenAndServe("localhost:7070", nil))
	}()
	println("=========Test function")
	time.Sleep(time.Second * 3)

	TestPool()
	//TestLargeObjects()


	wait := make(chan bool)
	<-wait
}

//测试大量对象的时候组件系统和传统方式的性能差异
func TestLargeObjects() {
	//====================== IComponent
	runtime := Component.NewRuntime(Component.Config{
		ThreadPoolSize: 50,
	})

	root := Component.NewObject("root")
	runtime.Root().AddObject(root)
	for i := 0; i < 1000; i++ {
		o1 := Component.NewObject(strconv.Itoa(i))
		o1.AddComponent(&Hello{})
		root.AddObject(o1)
	}

	t1 := time.Now()
	for i := 0; i < 10; i++ {
		runtime.Update(float32(1))
	}
	elapsed1 := time.Since(t1)
	println("component:", elapsed1)

	//========================== traditional
	tasklist := make([]*Hello, 1000)
	for i := 0; i < 1000; i++ {
		tasklist = append(tasklist, &Hello{})
	}
	wg := sync.WaitGroup{}
	taskqueue := make(chan int, 50)
	for j := 0; j < 50; j++ {
		go func() {
			for {
				index := <-taskqueue
				tasklist[index].Hello(strconv.Itoa(index))
				wg.Done()
			}
		}()
	}
	wg.Add(10000)
	t2 := time.Now()
	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {
			taskqueue <- j
		}
	}
	wg.Wait()
	elapsed2 := time.Since(t2)
	println("traditional:", elapsed2)
}

//测试使用pool和不使用pool的性能差异
//高负载时使用pool效果更好
func TestPool() {
	tasklist := make([]*Hello, 100000)
	for i := 0; i < 100000; i++ {
		tasklist = append(tasklist, &Hello{})
	}

	//====================== pool
	pool := threadpool.New()
	pool.MaxThreads = 50

	t1 := time.Now()
	wg1 := sync.WaitGroup{}
	wg1.Add(100000)
	for i := 0; i < 100000; i++ {

		pool.Run(func() {
			tasklist[i].Hello(strconv.Itoa(1))
			wg1.Done()
		})

	}

	wg1.Wait()
	elapsed1 := time.Since(t1)
	println("pool:", elapsed1)

	//========================== traditional

	t2 := time.Now()
	wg := sync.WaitGroup{}
	wg.Add(100000)
	for j := 0; j < 50; j++ {
		go func() {
			for i := 0; i < 2000; i++ {
				tasklist[i].Hello(strconv.Itoa(2))
				wg.Done()
			}

		}()
	}
	wg.Wait()
	elapsed2 := time.Since(t2)
	println("traditional:", elapsed2)

}
