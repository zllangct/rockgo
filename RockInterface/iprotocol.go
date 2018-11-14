package RockInterface

type IClientProtocol interface {
	OnConnectionMade(fconn Iclient)
	OnConnectionLost(fconn Iclient)
	StartReadThread(fconn Iclient)
	InitWorker(int32)
	AddRpcRouter(interface{})
	GetMsgHandle() Imsghandle
	GetDataPack() Idatapack
}

type IServerProtocol interface {
	OnConnectionMade(fconn ISession)
	OnConnectionLost(fconn ISession)
	StartReadThread(fconn ISession)
	InitWorker(int32)
	AddRpcRouter(interface{})
	GetMsgHandle() Imsghandle
	GetDataPack() Idatapack
}
