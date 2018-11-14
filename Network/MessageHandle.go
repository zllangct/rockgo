package Network

/*
	find msg api
*/
import (
	"fmt"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/utils"
	"reflect"
	"time"
	"runtime/debug"
	"sync"
)

type API struct {
	lock *sync.Mutex
	Route map[uint32]func(*PkgAll)
	ID2Message *map[reflect.Type]uint32
}

func (this *API)BaseInit(id2Message *map[reflect.Type]uint32)  {
	this.lock=&sync.Mutex{}
	this.Route= map[uint32]func(*PkgAll){}
	this.ID2Message=id2Message
}

func (this *API)On(msg interface{},handler func(*PkgAll))  {
	this.lock.Lock()
	defer this.lock.Unlock()

	if index,ok:=(*this.ID2Message)[reflect.TypeOf(msg)];ok {
		this.Route[index] = handler
	}
}

type MsgHandle struct {
	PoolSize  int32
	TaskQueue []chan *PkgAll
	Apis      map[uint32]reflect.Value

}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		PoolSize:  utils.GlobalObject.PoolSize,
		TaskQueue: make([]chan *PkgAll, utils.GlobalObject.PoolSize),
		Apis:      make(map[uint32]reflect.Value),
	}
}

//一致性路由,保证同一连接的数据转发给相同的goroutine
func (this *MsgHandle) DeliverToMsgQueue(pkg interface{}) {
	data := pkg.(*PkgAll)
	//index := rand.Int31n(utils.GlobalObject.PoolSize)
	index := data.Fconn.GetSessionId() % uint32(utils.GlobalObject.PoolSize)
	taskQueue := this.TaskQueue[index]
	logger.Debug(fmt.Sprintf("add to ConnPool : %d", index))
	taskQueue <- data
}

func (this *MsgHandle) DoMsgFromGoRoutine(pkg interface{}) {
	data := pkg.(*PkgAll)
	go func() {
		if f, ok := this.Apis[data.Pdata.MsgId]; ok {
			//存在
			st := time.Now()
			f.Call([]reflect.Value{reflect.ValueOf(data)})
			logger.Debug(fmt.Sprintf("Api_%d cost total time: %f ms", data.Pdata.MsgId, time.Now().Sub(st).Seconds()*1000))
		} else {
			logger.Error(fmt.Sprintf("not found api:  %d", data.Pdata.MsgId))
		}
	}()
}

//func (this *MsgHandle) AddRouter(router RockInterface{}) {
//	value := reflect.ValueOf(router)
//	tp := value.Type()
//	for i := 0; i < value.NumMethod(); i += 1 {
//		name := tp.Method(i).Name
//		k := strings.Split(name, "_")
//		index, err := strconv.Atoi(k[1])
//		if err != nil {
//			panic("error api: " + name)
//		}
//		if _, ok := this.Apis[uint32(index)]; ok {
//			//存在
//			panic("repeated api " + string(index))
//		}
//		this.Apis[uint32(index)] = value.Method(i)
//		logger.Info("add api " + name)
//	}
//}
func (this *MsgHandle) AddRouter(router interface{}) {
	value := reflect.ValueOf(router)
	value.MethodByName("Init").Call(nil)
	routeMap:= value.Elem().FieldByName("Route")
	keys := routeMap.MapKeys()

	for i, v := range keys {
		index:=uint32(v.Uint())
		if _, ok := this.Apis[index]; ok {
			panic("repeated api " + routeMap.MapIndex(v).String())
		}
		this.Apis[index] = routeMap.MapIndex(v)
		logger.Info("add api " + value.Type().Method(i).Name)
	}
}

func (this *MsgHandle)HandleError(err interface{}){
        if err != nil{
               debug.PrintStack()
        }
}

func (this *MsgHandle) StartWorkerLoop(poolSize int) {
	if utils.GlobalObject.IsThreadSafeMode(){
		//线程安全模式所有的逻辑都在一个goroutine处理, 这样可以实现无锁化服务
		this.TaskQueue[0] = make(chan *PkgAll, utils.GlobalObject.MaxWorkerLen)
		go func(){
			logger.Info("init thread mode workpool.")
			for{
				select {
				case data := <- this.TaskQueue[0]:
					if f, ok := this.Apis[data.Pdata.MsgId]; ok {
						//存在
						st := time.Now()
						//f.Call([]reflect.Value{reflect.ValueOf(data)})
						utils.XingoTry(f, []reflect.Value{reflect.ValueOf(data)}, this.HandleError)
						logger.Debug(fmt.Sprintf("Api_%d cost total time: %f ms", data.Pdata.MsgId, time.Now().Sub(st).Seconds()*1000))
					} else {
						logger.Error(fmt.Sprintf("not found api:  %d", data.Pdata.MsgId))
					}
				case delaytask := <- utils.GlobalObject.GetSafeTimer().GetTriggerChannel():
					delaytask.Call()
				}
			}
		}()
	}else{
		for i := 0; i < poolSize; i += 1 {
			c := make(chan *PkgAll, utils.GlobalObject.MaxWorkerLen)
			this.TaskQueue[i] = c
			go func(index int, taskQueue chan *PkgAll) {
				logger.Info(fmt.Sprintf("init thread ConnPool %d.", index))
				for {
					data := <-taskQueue
					//can goroutine?
					if f, ok := this.Apis[data.Pdata.MsgId]; ok {
						//存在
						st := time.Now()
						//f.Call([]reflect.Value{reflect.ValueOf(data)})
						utils.XingoTry(f, []reflect.Value{reflect.ValueOf(data)}, this.HandleError)
						logger.Debug(fmt.Sprintf("Api_%d cost total time: %f ms", data.Pdata.MsgId, time.Now().Sub(st).Seconds()*1000))
					} else {
						logger.Error(fmt.Sprintf("not found api:  %d", data.Pdata.MsgId))
					}
				}
			}(i, c)
		}
	}
}
