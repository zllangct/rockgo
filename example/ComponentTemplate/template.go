package main

import (
	"github.com/zllangct/RockGO/actor"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/config"
	"github.com/zllangct/RockGO/ecs"
	"github.com/zllangct/RockGO/logger"
	"reflect"
	"sync"
)

type TemplateComponent struct {
	ecs.ComponentBase              //Component 基类 必须继承
	Actor.ActorBase                //继承Actor基类，不使用actor模式时，不必继承
	locker            sync.RWMutex //锁
	member1           int          //成员变量1
	member2           string       //成员变量2
}

//指定该component是否是唯一，或唯一类型
//UNIQUE_TYPE_NONE ：不唯一
//UNIQUE_TYPE_LOCAL ：在parent object上面唯一
//UNIQUE_TYPE_GLOBAL ：在整个节点上面唯一
func (this *TemplateComponent) IsUnique() int {
	return ecs.UNIQUE_TYPE_GLOBAL
}

//指定该组件的依赖组件
func (this *TemplateComponent) GetRequire() map[*ecs.Object][]reflect.Type {
	requires := make(map[*ecs.Object][]reflect.Type)
	requires[this.Root()] = []reflect.Type{
		reflect.TypeOf(&config.ConfigComponent{}), //依赖根对象拥有ConfigComponent组件
		reflect.TypeOf(&Cluster.NodeComponent{}),  //依赖根对象拥有NodeComponent组件
	}
	/*
		requires[对象1] = []reflect.Type{
			reflect.TypeOf(&组件1{}),		//依赖根对象拥有组件1
		}
	*/
	return requires
}

/*
	组件的事件系统，选择有需求的事件，不需要时尽量删除空的接口实现
*/

//Awake 事件
func (this *TemplateComponent) Awake(context *ecs.Context) {
	//注册actor 消息处理函数
	this.AddHandler("HelloMessage", this.onHello)
}

//Start 事件
func (this *TemplateComponent) Start(ctx *ecs.Context) {

}

//Update 事件
func (this *TemplateComponent) Update(ctx *ecs.Context) {

}

//Destroy 事件
func (this *TemplateComponent) Destroy(ctx *ecs.Context) {

}

//组件的序列化
func (this *TemplateComponent) Serialize() (interface{}, error) {
	return nil, nil
}

//组件反序列化
func (this *TemplateComponent) Deserialize(data interface{}) error {
	return nil
}

//自定义函数
func (this *TemplateComponent) CustomFunction(role string, reply bool) error {
	return nil
}

func (this *TemplateComponent) onHello(message *Actor.ActorMessageInfo) error {
	logger.Debug(message.Message.Service)
	return nil
}

//根对象有nodeComponent 时，可定义RPC 调用函数，reply参数必须为指针
func (this *TemplateComponent) TestRPC(args string, reply *string) error {
	*reply = "hello" //注意reply的赋值方式
	return nil
}
