# RockGO

&emsp;&emsp;基于ECS(Entity component System)构建的分布式游戏服务端框架，同时提供Actor模型，目标是致力于快速搭建轻量、高性能、高可用的
分布式游戏后端，以及其他分布式后端应用。  

&emsp;&emsp;组件化的构架使得开发者不需要任何改动就能轻松实现一站式开发，分布式部署，规避分布式调试的困难，提升开发效率。框架提供udp、
tcp、websocket常用网络协议，同时提供优雅的协议接口，可让开发者轻松实现其他如kcp等网络协议的定制。

### Installation：
    $ go get -u github.com/zllangct/RockGO

### Quick start：

当然用最简单的hello world告诉你一切都是如此的简单，
详细代码参见example： [SingleNode](https://github.com/zllangct/RockGO/tree/master/example/SingleNode)  ，更多的用法，参见example目录。

Config：
```
    {
        "MasterAddress": "127.0.0.1:6666",        //中心服务节点地址
        "LocalAddress": "127.0.0.1:6666",         //本节点服务地址
        "AppName": "defaultApp",                  //本节点服务的App名称
        "Role": [                                 //本节点的服务角色，角色可理解为按服务内容分服
            "master"                              //此处，默认为中心服务节点
        ],
        "NodeDefine": {                           //节点定义
            "node_gate": {                        //程序启动参数中，附加节点名参数可覆盖上面的默认
                "LocalAddress": "0.0.0.0:6601",   //默认参数，eg：go run main.go -node node_gate
                "Role": [                         //此时启动的便是node_gate对应的服务内容
                    "gate"                        //节点服务内容可自由搭配，单服或者分布式的选择只
                ]                                 //需要修改配置文件，不需要修改任何一行代码，做到
            },                                    //单站式开发，随心部署
            "node_login": {
                "LocalAddress": "0.0.0.0:6602",
                "Role": [
                    "login"
                ]
            }
        },        
        "NetConnTimeout": 9000,                 //外网连接心跳超时间隔，单位毫秒
        "NetListenAddress": "0.0.0.0:5555",     //外网服务端口
    }
```
Server：  
```
    package main
    
    import (
    	"flag"
    	"fmt"
    	"github.com/zllangct/RockGO"
    	"github.com/zllangct/RockGO/component"
    	"github.com/zllangct/RockGO/gate"
    	"github.com/zllangct/RockGO/logger"
    )

    var Server *RockGO.Server
    
    func main()  {  
        //初始化服务节点
    	Server = RockGO.DefaultServer()  
    	
    	/*
    	    添加组件组
    	    添加网关组件（DefaultGateComponent）后，此服务节点拥有网关的服务能力。
    	    同理，添加其他组件，如登录组件（LoginComponent）后，拥有登录的服务内容。
    	*/
    	Server.AddComponentGroup("gate",[]Component.IComponent{&gate.DefaultGateComponent{}})
	
    	//开始服务
    	Server.Serve()
    }  
```
Client：
```
//默认网关采用websocket协议，可以使用任务websocket客户点连接测试。example中提供一个laya客户端供测试使用。
```
### Feature：

####1. ECS（Entity Component System）构架：    

&emsp;&emsp;ECS全称Entity-Component-System（实体-组件-系统），是基于组合优于继承，即将不变的部分使用继承以方便复用， 
将多变的部分用组合来方便拓展，是按照这种原则的一种设计模式。当然这种设计模式带来了许多比OOP更容易容实现
的特性，当然ECS并不是抛弃OOP，这是一个度的问题。ECS的特点和优势，暴雪在分享《守望先锋》的ECS服务构架的文
章中已经阐述比较详细[（中文译文戳这里）](http://gad.qq.com/article/detail/28682)，也可以看看云风博客中的
讨论  [（继续戳）](https://blog.codingnow.com/2017/06/overwatch_ecs.html)，此处都不赘述了。在服务端程序中
的优势，这里简要概括几点：    
 
#####(1). 自由组合：
        Component尽量按照原子性原则设计，只要有合适的功能拆分粒度，服务端可以随性所欲的组合，这在分布式构架
    中至关重要，关系到服务端节点间功能的拆分组合。比如中心服务节点担任了游戏大厅和注册登录两项服务，随着用户
    规模的扩大，中心节点需要把注册登录拆分出来单独成为节点。在ECS构架中，这将是非常容易的一件事，很少或者
    甚至不需要任何修改一行功能代码就能实现，因为节点本身就是功能组件的组合，分服不过是根据配置文件重新组合一
    次。这也微服务的思想走在了一起，微服务往往在服务间的调用走的是网络调用，但ECS构架下可以轻松的实现网络本地
    调用的自主切换。单服就是真正的单服，本地调用，减少不必要的网络消耗，而不是多个服务部署在了同一台物理机上。
        
#####(2). 热插拔：
        ECS构架下，实体和组件都能运行时添加删除，框架提供了Initialize、Awake、Start、Update、Destroy 内置系统，
    这些内置系统能保证每一个功能组件，或者组件组合所需的完整生命周期。
#####(3). 易拓展：
        功能的拓展，大多情况是新组件，新系统的设计，耦合性极低。
#####(4). 更优雅的停机：
        所以对象都有完整的生命周期，Destroy系统可以保证每个对象在销毁时，得到正确的处理，实现更优雅的停机处理。
#####(5). 序列化：
        所有组件，实现持久化接口之后，都具备序列化的能力。配合优雅的停机，很容易实现停机后的恢复。
#####(7). 高性能
        每个系统都不会遍历所有的对象，只会过滤出感兴趣的组件，专人专事，效率集中，减少调用，例如不需要Update处理的组件，便
    不用实现Update接口，Update系统将不会遍历此组件。 
#####(8). 方便调试
        分布式调试的麻烦，这里并不存在，单机开发，随心大胆的断点，最后仅仅是一个配置参数完成分布式部署。
#####(9). 想到了再添加 ... ...      
 
####2. Actor 模式：
&emsp;&emsp;有空再详细写，反正知道对方的ActorID，无论他在哪儿，无论活在那个节点，Tell() 都能告诉他。框架内自主实现本地调用和RPC
调用的分流。Actor模式的特性与优势，看官自行G或者B。
####3. RPC：
&emsp;&emsp;RPC的性能与可靠性是分布式系统中最重要的一环，游戏要的是效率，服务治理方面够用就行，所以并未选择市面上流行的服务治理型
的RPC调用框架，并且游戏构架本身就具备了服务治理的能力，但RPC框架自带的治理能力又不足以支撑游戏需求，所以略显多余，大体量的框架
复杂的环境配置大大的增加了维护、定制、学习、部署、迁移的难度，尤其对中小型开发者不友好。既要满足中小型项目需要的简易性，
又具有大型项目的扩展性，框架选用了go自带的rpc框架，其性能有目共睹，测评参照这里（[戳我就行](https://github.com/smallnest/rpcx)）,
基于net/rpc作了轻量化定制，添加了必要的功能特性：
######&emsp;&emsp;&emsp;&emsp;(1). 超时机制 
######&emsp;&emsp;&emsp;&emsp;(2). 心跳检测
######&emsp;&emsp;&emsp;&emsp;(3). 断线重连
######&emsp;&emsp;&emsp;&emsp;(4). 单向调用
######&emsp;&emsp;&emsp;&emsp;(5). 客户端回调

####4. 网络协议：
&emsp;&emsp;框架支持的网络协议有：TCP、UDP、Websocket，封包格式如下：     

| 协议名称 | 协议格式 | 数据类型 |
| :------:| :------: | :------: |
| TCP | Length-[Type-Data] | 二进制 |
| UDP | Length-[Session-Type-Data] | 二进制 |
| Websocket | Type-Data | 二进制 |
####5. Actor 模式：

####6. 计划任务：

####7. 写在后面：
&emsp;&emsp;  致谢：  [gin — gin-gonic](https://github.com/gin-gonic/gin)、[websocket—gorilla](https://github.com/gorilla/websocket)
、[go-component—shadowmint](https://github.com/shadowmint/go-component)、[TarsGo—TarsCloud](https://github.com/TarsCloud/TarsGo)   
&emsp;&emsp;  喜欢的朋友给个星星~  

&emsp;&emsp;【RockGO交流群①】（已满，请加②群）






