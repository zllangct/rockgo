package sys_rpc

import (
	"github.com/zllangct/RockGO/clusterOld"

	"github.com/zllangct/RockGO/clusterserver"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/utils"
)

type RootRpc struct {
}

/*
子节点连上来的通知
*/
func (this *RootRpc) TakeProxy(request *clusterOld.RpcRequest) {
	name := request.Rpcdata.Args[0].(string)
	logger.Info("child node " + name + " connected to " + utils.GlobalObject.Name)
	request.Fconn.SetProperty("child",name)
	//加到childs并且绑定链接connetion对象
	clusterserver.GlobalClusterServer.AddChild(name, request.Fconn)
	//通知child,连接准备成功
	child,err:= clusterserver.GlobalClusterServer.ChildsMgr.GetChild(name)
	if err != nil{
		logger.Info("AddChild filed"+err.Error())
	}
	child.CallChildNotForResult("ConnectOK",utils.GlobalObject.Name)
	//触发子节点链接成功事件
	utils.GlobalObject.OnChildNodeConnected(name,request.Fconn)
}

/*
添加工作conn ConnPool
*/
func (this *RootRpc)AddChildConnPool(request *clusterOld.RpcRequest)  {
	if !utils.GlobalObject.MultiConnMode{
		return
	}
	name := request.Rpcdata.Args[0].(string)
	request.Fconn.SetProperty("child",name)
	child,err:= clusterserver.GlobalClusterServer.ChildsMgr.GetChild(name)
	if err != nil{
		logger.Info("AddChildConnPool filed "+err.Error())
	}else{
		request.Fconn.SetProperty("remote",name)
		err:=child.AddWorkConn(request.Fconn)
		if err !=nil{
			logger.Error("AddWorkConn filed "+err.Error())
		}
	}
}
