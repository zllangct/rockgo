package main

/*
	1、
	1、不同节点部署在不同的物理机上面，
	{
		"MasterAddress": "127.0.0.1:6666",     // master节点地址，内网填写局域网地址，非局域网需要填写完整外网地址
		"LocalAddress": "0.0.0.0:6666",		   // 任意可用端口
		"AppName": "defaultApp",
		"Role": [
			"master"						  //默认角色，当不指定角色时为master，在此设置角色
		],
		"NodeDefine": {},					  //不需要定义

		//以下是其他配置
		"ReportInterval": 3000,
		"RpcTimeout": 9000,
		"RpcCallTimeout": 5000,
		"RpcHeartBeatInterval": 3000,
		"IsLocationMode": true,
		"NetConnTimeout": 9000,
		"NetListenAddress": "0.0.0.0:5555",
		"IsActorModel": true
	}
	2、不同节点分开部署在同一物理机，由于同一台物理机节点监听端口不能相同，所以分开配置, 分别指定默认角色，或者
		使用eg：server.exe -node node_gate方式启动。

	{
		"MasterAddress": "127.0.0.1:6666",
		"LocalAddress": "127.0.0.1:6666",
		"AppName": "defaultApp",
		"Role": [
			"single"
		],
		"NodeDefine": {
			"node_actor_location": {
				"LocalAddress": "0.0.0.0:6604",
				"Role": [
					"actor_location"
				]
			},
			"node_gate": {
				"LocalAddress": "0.0.0.0:6601",
				"Role": [
					"gate"
				]
			},
			"node_location": {
				"LocalAddress": "0.0.0.0:6603",
				"Role": [
					"location"
				]
			},
			"node_login": {
				"LocalAddress": "0.0.0.0:6602",
				"Role": [
					"login"
				]
			},
			"node_master": {
				"LocalAddress": "0.0.0.0:6666",
				"Role": [
					"master"
				]
			},
			"node_room": {
				"LocalAddress": "0.0.0.0:6605",
				"Role": [
					"room"
				]
			},
		},
		"ReportInterval": 3000,
		"RpcTimeout": 9000,
		"RpcCallTimeout": 5000,
		"RpcHeartBeatInterval": 3000,
		"IsLocationMode": true,
		"NetConnTimeout": 9000,
		"NetListenAddress": "0.0.0.0:5555",
		"IsActorModel": true
	}
*/


