# RockGO

&emsp;&emsp;基于ECS(Entity component System)构建的分布式游戏服务端框架，目标是致力于快速搭建轻量、高性能、高可用的分布式游戏后端，
以及其他分布式后端应用。  
&emsp;&emsp;组件化的构架使得开发者不需要任何改动就能轻松实现一站式开发，分布式部署，规避分布式调试的困难，提升开发效率。框架提供udp、
tcp、websocket常用网络协议，同时提供优雅的协议接口，可让开发者轻松实现其他如kcp等网络协议的定制。


### 开发便签：
    

### 任务列表：
&emsp;&emsp;1、rpc 本地调用，当调用者和服务者都是本地时，不走网络途径，实现自动分流，在node component中
通过ip:port判断是否为本地调用，调用rpc server的local call 方法实现。
