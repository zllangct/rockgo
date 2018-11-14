#ZGO

框架主要用于分布式在线游戏服务端，核心模块包含：网络模块，RPC模块，集群管理模块，以及游戏后端常用工具。

一、Helloworld

先通过一个helloworld程序展示怎么开始使用这个框架。话不多说，先看代码：
<pre><code>
func main() {
    //初始化一个server
    s := zgo.NewXingoCluterServer("gate1", filepath.Join(dir, "Conf", "clusterconf.json"))
    //为这个server添加对应的逻辑模组
    s.AddModule("gate", nil, nil, &GateServer.GateRpc{})
    //运行server，开始服务
    s.StartClusterServer()
}
</code></pre>
简单的三行代码就可以使用zgo启动一个服务器节点，NewXingoCluterServer这个函数需要接受节点名称和配置文件地址两个参数，其中节点名称需和配置文件中名称一致。
    AddModule这个函数接受四个参数，第一个为节点名字，第二个netModule，即外网模块，负责处理客户端的消息。第三个httpModule，于第一个一致，只不过是走http方式。
    第四个参数为rpcModule，包含节点内网的rpc逻辑。
    如果某个节点只有一个rpc逻辑，不对外网服务那么netModule和httpModule 为nil。
    比如网管节点需要有netModule和rpcModule，web后台管理节点则需要httpModule和rpcModule。
    
二、框架目录介绍

1、cluster：分布式集群中节点定义，RPC的实现。

2、cluterserver：分布式中心服务器和节点服务器的实现

3、db：MongoDB的链接实现，其他数据库可以自己添加

4、Docments：文档

5、fnet：外网模块，处理客户端链接，数据解析

6、fserver：外网服务端实现

7、iface：接口

8、logger：日志模块

9、pool：对象池

10、sys_rpc：rpc相关实现

11、telnetcmd：telnet命名模块

12、timer：计时器

13、utils：工具函数

三、框架设计思路

框架目标是打造一个轻量级、高并发、高可用的分布式服务端程序。既然是分布式，那么即是对整个服务端程序按照功能模块
进行拆分，每个节点由一个或几个功能模块组成，每一个功能模块也可以存在多个节点，所有的节点共同构成完整的分布式集群。
    
节点分为masterNode和LogicNode，即中心节点和逻辑节点。

masterNode 负责整个集群的管理工作，不做任何逻辑处理。实现节点的注册与发现，检测各节点的健康状态。

LogicNode 服务端的业务逻辑节点，完成拆分后的某个功能模块，比如登录模块。

节点间通过配置文件获取各节点地址，通过master获取节点的在线状态，然后各节点通过RPC交互。

至于节点分布需要根据业务逻辑去决定。业务逻辑尽量去中心化，消息路由通过一致性hash去完成。