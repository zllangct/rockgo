package main

import (
	"github.com/zllangct/RockGO/actor"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"reflect"
	"sync"
)

type TemplateComponent struct {
	Component.Base 					//Component 基类 必须继承
	locker        sync.RWMutex		//锁
	member1			int				//成员变量1
	member2	        string			//成员变量2
}

//指定该component是否是唯一，或唯一类型
//UNIQUE_TYPE_NONE ：不唯一
//UNIQUE_TYPE_LOCAL ：在parent object上面唯一
//UNIQUE_TYPE_GLOBAL ：在整个节点上面唯一
func (this *TemplateComponent) IsUnique() int {
	return Component.UNIQUE_TYPE_GLOBAL
}

//指定该组件的依赖组件
func (this *TemplateComponent) GetRequire() map[*Component.Object][]reflect.Type {
	requires := make(map[*Component.Object][]reflect.Type)
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),		//依赖根对象拥有ConfigComponent组件
		reflect.TypeOf(&Cluster.NodeComponent{}),		//依赖根对象拥有NodeComponent组件
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
func (this *TemplateComponent) Awake()error {
	return nil
}

//Start 事件
func (this *TemplateComponent) Start(ctx *Component.Context) {

}

//Update 事件
func (this *TemplateComponent) Update(ctx *Component.Context) {

}

//Destroy 事件
func (this *TemplateComponent) Destroy()error {
	return nil
}

//组件的序列化
func (this *TemplateComponent)Serialize() (interface{}, error){
	return nil,nil
}

//组件反序列化
func (this *TemplateComponent)Deserialize(data interface{}) error{
	return nil
}

//自定义函数
func (this *TemplateComponent)CustomFunction(role string,reply bool)error  {
	return nil
}

/*
	扩充，当有其他组件配合时，可扩充功能，同时，组件本身的扩展，可依赖于其他组件，或者继承其他基类
		比如，可继承network.ApiBase基类，使之拥有协议API解析能力，开发者可开发自定义功能类，也可配合
		其他依赖组件，完成功能，比如下方的，配合Actor组件处理actor事件，配合nodeComponent组件完成rpc
		的直接调用。
*/

//父对象有ActorComponent时，可以定义Actor消息处理函数
func (this *TemplateComponent)GetMessageHandler()map[string]func(message *Actor.ActorMessageInfo)  {
	//该map建议在Awake中初始化，避免反复创建，此处仅便宜行事
	return  map[string]func(message *Actor.ActorMessageInfo){
		"hello":this.onHello,
	}
}

func (this *TemplateComponent)onHello(message *Actor.ActorMessageInfo) {

}

//根对象有nodeComponent 时，可定义RPC 调用函数，reply参数必须为指针
func (this *TemplateComponent)TestRPC(args string,reply *string) error {
	*reply = "hello"	//注意reply的赋值方式
	return nil
}












