# RockGO（go master分支泛型基本可以使用，本库使用泛型重构后发布）

重构 RockGO ECS engine: https://github.com/zllangct/ecs

&emsp;&emsp;基于ECS(Entity component System)构建的分布式游戏服务端框架，同时提供Actor模型，目标是致力于快速搭建轻量、高性能、高可用的
分布式游戏后端，以及其他分布式后端应用。  

&emsp;&emsp;组件化的构架使得开发者不需要任何改动就能轻松实现一站式开发，分布式部署，规避分布式调试的困难，提升开发效率。框架提供udp、
tcp、websocket常用网络协议，同时提供优雅的协议接口，可让开发者轻松实现其他如kcp等网络协议的定制。

### Installation：
    $ go get -u github.com/zllangct/RockGO

### Quick start：

当然用最简单的hello world告诉你一切都是如此的简单，
完整代码参见example： [SingleNode](https://github.com/zllangct/RockGO/tree/master/example/SingleNode)  ，更多的用法，参见example目录。

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
                "Role": [                         //分布式部署：
                    "login"                       //物理机一：go run main.go -node node_master
                ]                                 //物理机二：go run main.go -node node_gate
            },                                    //物理机三：go run main.go -node node_room
            "node_login_gate": {
                "LocalAddress": "0.0.0.0:6603",   //go run main.go -node node_gate_gate 这样启动的
                "Role": [                         //便是一个节点同时具备 login 和 gate 两个角色的节点
                    "login",
                    "gate" 
                ]
            }
            "node_single": {
                "LocalAddress": "0.0.0.0:6604",   //go run main.go -node single 这样启动的
                "Role": [                         //便是所有服务在同一节点，即单服模式，小负载
                    "login",                      //或者开发阶段使用，方便调试
                    "gate" 
                    "master",
                    "room",
                    "location"
                ]
            }
        },        
        "NetConnTimeout": 9000,                 //外网连接心跳超时间隔，单位毫秒
        "NetListenAddress": "0.0.0.0:5555",     //外网服务端口
    }
```
Server：  
```go
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
//默认网关采用websocket协议，可以使用任意websocket客户点连接测试。example中提供一个laya客户端供测试使用。
```
### Feature：

#### 1. ECS（Entity Component System）构架    

&emsp;&emsp;ECS全称Entity-Component-System（实体-组件-系统），是基于组合优于继承，即将不变的部分使用继承以方便复用， 
将多变的部分用组合来方便拓展，是按照这种原则的一种设计模式。当然这种设计模式带来了许多比OOP更容易容实现
的特性，当然ECS并不是抛弃OOP，这是一个度的问题。ECS的特点和优势，暴雪在分享《守望先锋》的ECS服务构架的文
章中已经阐述比较详细[（中文译文戳这里）](http://gad.qq.com/article/detail/28682)，也可以看看云风博客中的
讨论  [（继续戳）](https://blog.codingnow.com/2017/06/overwatch_ecs.html)，此处都不赘述了。在服务端程序中
的优势，这里简要概括几点：    
 
##### 1.1 自由组合
        Component尽量按照原子性原则设计，只要有合适的功能拆分粒度，服务端可以随性所欲的组合，这在分布式构架
    中至关重要，关系到服务端节点间功能的拆分组合。比如中心服务节点担任了游戏大厅和注册登录两项服务，随着用户
    规模的扩大，中心节点需要把注册登录拆分出来单独成为节点。在ECS构架中，这将是非常容易的一件事，很少或者
    甚至不需要任何修改一行功能代码就能实现，因为节点本身就是功能组件的组合，分服不过是根据配置文件重新组合一
    次。这也微服务的思想走在了一起，微服务往往在服务间的调用走的是网络调用，但ECS构架下可以轻松的实现网络本地
    调用的自主切换。单服就是真正的单服，本地调用，减少不必要的网络消耗，而不是多个服务部署在了同一台物理机上。
        
##### 1.2 热插拔
        ECS构架下，实体和组件都能运行时添加删除，框架提供了Initialize、Awake、Start、Update、Destroy 内置系统，
    这些内置系统能保证每一个功能组件，或者组件组合所需的完整生命周期。
##### 1.3 易拓展
        功能的拓展，大多情况是新组件，新系统的设计，耦合性极低。
##### 1.4 更优雅的停机
        所以对象都有完整的生命周期，Destroy系统可以保证每个对象在销毁时，得到正确的处理，实现更优雅的停机处理。
##### 1.5 序列化
        所有组件，实现持久化接口之后，都具备序列化的能力。配合优雅的停机，很容易实现停机后的恢复。
##### 1.7 高性能
        每个系统都不会遍历所有的对象，只会过滤出感兴趣的组件，专人专事，效率集中，减少调用，例如不需要Update处理的组件，便
    不用实现Update接口，Update系统将不会遍历此组件。 
##### 1.8 方便调试
        分布式调试的麻烦，这里并不存在，单机开发，随心大胆的断点，最后仅仅是一个配置参数完成分布式部署。
##### 1.9 想到了再添加 ... ...      
 
#### 2. Actor 模式
&emsp;&emsp;有空再详细写，反正知道对方的ActorID，无论他在哪儿，无论活在那个节点，Tell() 都能告诉他。框架内自主实现本地调用和RPC
调用的分流。Actor模式的特性与优势，看官自行G或者B。千万别问为什么有了ECS还能有Actor，并不冲突，ECS也是基于OOP实现的，所以Actor基于
ECS实现，而且比OOP实现Actor更简单。
#### 3. RPC
&emsp;&emsp;RPC的性能与可靠性是分布式系统中最重要的一环，游戏要的是效率，服务治理方面够用就行，所以并未选择市面上流行的服务治理型
的RPC调用框架，并且游戏构架本身就具备了服务治理的能力，但RPC框架自带的治理能力又不足以支撑游戏需求，所以略显多余，大体量的框架
复杂的环境配置大大的增加了维护、定制、学习、部署、迁移的难度，尤其对中小型开发者不友好。既要满足中小型项目需要的简易性，
又具有大型项目的扩展性，框架选用了go自带的rpc框架，其性能有目共睹，测评参照这里（[戳我就行](https://github.com/smallnest/rpcx)）,
基于net/rpc作了轻量化定制，添加了必要的功能特性：
###### &emsp;&emsp;&emsp;(1). 超时机制 &emsp;(2). 心跳检测&emsp;(3). 断线重连&emsp;(4). 单向调用&emsp;(5). 客户端回调

#### 4. 网络协议
##### 4.1 框架支持的网络协议
框架支持的网络协议有：TCP、UDP、Websocket、http，封包格式如下：     

| 协议名称 | 协议格式 | 长度  |  数据类型 |
| :------:| :------: | :------: |:------: |
| TCP | Length-[Type-Data] |4 - [ 4 - n ]| 二进制 |
| UDP | Length-[Session-Type-Data] |4 - [ 4 - 4 - n ]| 二进制 |
| Websocket | Type-Data |4 - n| 二进制 |       

&emsp;&emsp;http建议直接在网关组件中使用gin、fasthttp等http处理框架对应路由处理函数，http使用途中极有可能与页面有关，虽然
集成到上述方式路由中十分简单，但不建议这样处理，这样僵化了http的灵活性和丰富的功能特性。
##### 4.2 自定义网络协议
&emsp;&emsp;框架内可以很轻松的实现自定义协议的扩展，对原有协议都是非侵入的依赖，只需要简单封装，比如各种可靠UDP协议，KCP、UDT、
ENET等，只需要实现以下述接口：
```go
    //网络协议
    type ServerHandler interface {
        Listen() error                      //监听
        Handle() error                      //处理
    }
    //链接对象
    type Conn interface {
    	WriteMessage(messageType uint32, data []byte) error   //消息发送
    	Addr() string                                         //目标地址
    	Close() error                                         //关闭
    }
    //解包协议
    type Protocol interface {
        ParsePackage( []byte) (int, int)                            //包处理
    	ParseMessage( context.Context,  []byte)([]uint32,[]byte)    //消息处理
    }
   
```
&emsp;&emsp;网络协议的实现可参照TCP、UDP和Websocket。各种网络协议所提供链接对象可不相同，
需要实现连接对象接口进行统一，供Session使用。网络协议和链接对象接口的实现相对简单，不容
易产生歧义，其中解包协议参照TCP和Websocket用法：
```go
/* TCP LTD protocol 
	Length—（Type—Data） ，数据长度—（消息类型—消息体） 大小：  4 — （4 — n）
*/
type LtdProtocol struct{}

//完整包中解析出消息ID和数据部分
func (s *LtdProtocol) ParseMessage(ctx context.Context,data []byte)([]uint32,[]byte){
    mt := binary.BigEndian.Uint32(data[:4])
    return []uint32{mt}, data[4:]
}

//检查包是否接受完整
func (s *LtdProtocol) ParsePackage(buff []byte) (pkgLen, status int) {
    if len(buff) < 4 {
        return 0, PACKAGE_LESS
    }
    length := binary.BigEndian.Uint32(buff[:4])
    
    if length > 1048576000 || len(buff) > 1048576000 { // 1000MB
        return 0, PACKAGE_ERROR
    }
    if len(buff) < int(length) {
        return 0, PACKAGE_LESS
    }
    return int(length), PACKAGE_FULL
}


/* Websocket TD protocol
	Type—Data ，消息类型—消息体 大小：  4 — n
*/
type TdProtocol struct{}

//解析消息ID和消息数据
func (s *TdProtocol) ParseMessage(ctx context.Context,data []byte)([]uint32,[]byte){
    mt := binary.BigEndian.Uint32(data[:4])
    return []uint32{mt}, data[4:]
}

//websocket 自带粘包处理，此处无需手动处理
func (s *TdProtocol) ParsePackage(buff []byte) (pkgLen, status int) {
    return 0,0
}
```
#### 5. 消息序列化协议
##### 5.1 支持的序列化协议
###### &emsp;&emsp;&emsp;(1). Json &emsp;(2). ProtoBuf
##### 5.2 自定义序列化协议
&emsp;&emsp;自定义序列化协议需要实现以下接口：
```go
//消息解析协议
type MessageProtocol interface {
    Marshal( interface{})([]byte,error)         //序列化
    Unmarshal( []byte, interface{})error        //反序列化
}
```
#### 7. 业务接口
##### 7.1 路由规则 
&emsp;&emsp;框架提供消息的路由，无需用户手动对应消息的处理方法，框架根据函数自动判断是否为消息处理函数。需要满足以下条件
： 
###### &emsp;&emsp;(1). 结构体继承 ApiBase 
###### &emsp;&emsp;(2). 函数必须为结构体导出函数
###### &emsp;&emsp;(3). 函数必须为以下结构：func (this *XXX) FunctionName(sess *network.Session,message *MessageStruct)
###### &emsp;&emsp; 第一参数为 会话Session，第二参数为消息对应的结构体，框架会更加第二参数去判断处理对应的消息。
&emsp;&emsp;参见:
```go
//协议对应字典，推荐使用工具生成该文件，以便前后端对应准确
//稍后会提供相应工具，目前完成了protobuf 导出c# 和 golang 的协议对应
//完善之后会更新至本仓库，由于过于简单，客官可自行完成
// 原理：1）读取proto文件 2）提取消息名 3）按照同一序号生成c#、golang或者其他语言文件（字符串拼接）

var Testid2mt = map[reflect.Type]uint32{
    reflect.TypeOf(&TestMessage{}):1,
    reflect.TypeOf(&TestLogin{}):2,
    reflect.TypeOf(&PlayerInfo{}):3,
}

//消息定义
type TestMessage struct {
    Name string
}
type TestReply struct {
    Result bool
}

//接口组定义
type TestApi struct {
    network.ApiBase         //继承ApiBase
}

/*
    使用协议接口时，需先初始化，初始化时需传入定义的消息号对应字典
    以及所需的消息序列化组件，可轻易切换为protobuf，msgpack等其他序列化工具
*/
func NewTestApi() *TestApi  {
    r:=&TestApi{}
    r.Instance(r).SetMT2ID(Testid2mt).SetProtocol(&MessageProtocol.JsonProtocol{})
    return r
}

//协议接口1  Hello,框架会自动判断TestMessage类型消息，自动路由至此函数处理
func (this *TestApi)Hello(sess *network.Session,message *TestMessage) {
    //打印消息
    println(fmt.Sprintf("Hello,%s", message.Name))

    //回复消息
    res:=&TestReply{
        Result:true,
    }
    this.Reply(sess,res)
}

//协议接口2  other，同理，该函数处理 Other 类型消息
func (this *TestApi) Other(sess *network.Session,message *Other) {
	......
}
```
##### 7.1 自定义路由 
&emsp;&emsp;当然用户可以不用使用框架自带的消息路由方法，可以实现NetAPI接口自定义消息路由规则：
```go
type NetAPI interface {
        Init()                                                                        //初始化
	Route(*Session, uint32, []byte)	                                              //反序列化并路由到api处理函数
	Reply(session *Session,message interface{})error                              //序列化消息并发送至客户端
}
```
#### 8. Example列表

###### &emsp;&emsp;[组件定义范例](https://github.com/zllangct/RockGO/tree/master/example/ComponentTemplate)
###### &emsp;&emsp;[Laya测试客户端](https://github.com/zllangct/RockGO/tree/master/example/DebugClient)
###### &emsp;&emsp;[Golang Websocket测试客户端](https://github.com/zllangct/RockGO/tree/master/example/GoDebugClient)
###### &emsp;&emsp;[分布式部署配置范例](https://github.com/zllangct/RockGO/tree/master/example/MultiNodeServer)
###### &emsp;&emsp;[ECS单节点范例](https://github.com/zllangct/RockGO/tree/master/example/SingleNode)
###### &emsp;&emsp;[ECS+Actor单节点范例](https://github.com/zllangct/RockGO/tree/master/example/SingleNodeWithActor)
###### &emsp;&emsp;[TCPUDP服务端客户端范例](https://github.com/zllangct/RockGO/tree/master/example/TcpOrUdpServer)
###### &emsp;&emsp;[Websocket服务端范例](https://github.com/zllangct/RockGO/tree/master/example/WebsocketServer)

#### 8. 计划任务
###### &emsp;&emsp;(1). 完善现有example
###### &emsp;&emsp;(2). 提供房间类游戏大厅、房间模型（就是你想的棋牌）
###### &emsp;&emsp;(3). 增加Laya 客户端范例
###### &emsp;&emsp;(4). 增加Unity 帧同步客户端范例
###### &emsp;&emsp;(5). 增加KCP协议支持
###### &emsp;&emsp;(6). 后台管理页面、数据统计页面
###### &emsp;&emsp;(7). 节点平滑升级
###### &emsp;&emsp;(8). 提供protobuf 前后端协议自动化对应工具
###### &emsp;&emsp;。。。。。。
#### 9. 写在后面
&emsp;&emsp;  致谢：  [gin — gin-gonic](https://github.com/gin-gonic/gin)、[websocket—gorilla](https://github.com/gorilla/websocket)
、[go-component—shadowmint](https://github.com/shadowmint/go-component)、[TarsGo—TarsCloud](https://github.com/TarsCloud/TarsGo)   
&emsp;&emsp;  喜欢的朋友给个星星~  

&emsp;&emsp;【RockGO交流群①】（已满，请加②群）






