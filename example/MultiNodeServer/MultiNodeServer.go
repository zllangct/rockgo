package main

/*
	相对于SingleNode 范例，多节点多角色模式，仅仅是部署上的不同，所以单服和多服开发上没有任何区别
	仅仅部署时，配置文件参数不同。在部署时，ClusterConfig.json 的配置文件中，Role 属性 为该节点所
	担任的角色类型，可任意搭配，但注意，master节点必须唯一，不是master节点，或master节点不担任其他
    角色时可省略child 角色，系统自动添加。master 仅做节点信息收集，无任何逻辑，且无重要信息，夯机危
	害性小，一般情况不做冗余配置。若有需要，开发者可自行根据类似Raft等算法，设置备用节点。

	单服的集群配置文件：
	{
		"MasterAddress": "127.0.0.1:6666",
		"LocalAddress": "127.0.0.1:6666",
		"AppName": "defaultApp",
		"Role": [
			"master","child","login","gate"
		],
		"Group": [],
		"ReportInterval": 3000,
		"RpcTimeout": 9000,
		"RpcCallTimeout": 5000,
		"RpcHeartBeatInterval": 3000,
		"NetConnTimeout": 9000,
		"NetListenAddress": "0.0.0.0:5555"
	}
	分服后，Master 节点：
	{
		"MasterAddress": "127.0.0.1:6666",
		"LocalAddress": "127.0.0.1:6666",
		"AppName": "defaultApp",
		"Role": [
			"master"
		],
		"Group": [],
		"ReportInterval": 3000,
		"RpcTimeout": 9000,
		"RpcCallTimeout": 5000,
		"RpcHeartBeatInterval": 3000,
		"NetConnTimeout": 9000,
		"NetListenAddress": "0.0.0.0:5555"
	}
	分服后，Gate 网关节点：
	{
		"MasterAddress": "127.0.0.1:6666",
		"LocalAddress": "127.0.0.1:6666",
		"AppName": "defaultApp",
		"Role": [
			"child","gate"
		],
		"Group": [],
		"ReportInterval": 3000,
		"RpcTimeout": 9000,
		"RpcCallTimeout": 5000,
		"RpcHeartBeatInterval": 3000,
		"NetConnTimeout": 9000,
		"NetListenAddress": "0.0.0.0:5555"
	}
	分服后，Login  登录节点：
	{
		"MasterAddress": "127.0.0.1:6666",
		"LocalAddress": "127.0.0.1:6666",
		"AppName": "defaultApp",
		"Role": [
			"child","login"
		],
		"Group": [],
		"ReportInterval": 3000,
		"RpcTimeout": 9000,
		"RpcCallTimeout": 5000,
		"RpcHeartBeatInterval": 3000,
		"NetConnTimeout": 9000,
		"NetListenAddress": "0.0.0.0:5555"
	}
*/


