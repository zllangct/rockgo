package sys_rpc

import (
	"github.com/zllangct/RockGO/clusterOld"
	"github.com/zllangct/RockGO/clusterserver"
	"github.com/zllangct/RockGO/logger"
)

type MasterRpc struct {
}

func (this *MasterRpc) TakeProxy(request *clusterOld.RpcRequest) (response map[string]interface{}) {
	response = make(map[string]interface{}, 0)
	name := request.Rpcdata.Args[0].(string)
	logger.Info("node " + name + " connected to master.")

	if child,err:=clusterserver.GlobalMaster.Childs.GetChild(name);err==nil {
		logger.Info("node:"+name+" already  exist")
		response["reason"]="this node :"+name+" is already exist,dont open repeated."
		go clusterserver.GlobalMaster.CheckChildAlive(child)
		return
	}

	//加到childs并且绑定链接connetion对象
	clusterserver.GlobalMaster.AddNode(name, request.Fconn)

	//返回需要链接的父节点
	remotes, err := clusterserver.GlobalMaster.Cconf.GetRemotesByName(name)
	if err == nil {
		roots := make([]string, 0)
		for _, r := range remotes {
			if _, ok := clusterserver.GlobalMaster.OnlineNodes[r]; ok {
				//父节点在线
				roots = append(roots, r)
			}
		}
		response["roots"] = roots
	}
	//通知当前节点的子节点链接当前节点
	for _, child := range clusterserver.GlobalMaster.Childs.GetChilds() {
		//遍历所有子节点,观察child节点的父节点是否包含当前节点
		remotes, err := clusterserver.GlobalMaster.Cconf.GetRemotesByName(child.GetName())
		if err == nil {
			for _, rname := range remotes {
				if rname == name {
					//包含，需要通知child节点连接当前节点
					//rpc notice
					child.CallChildNotForResult("RootTakeProxy", name)
					break
				}
			}
		}
	}
	return
}

//主动通知master 节点掉线
func (this *MasterRpc) ChildOffLine(request *clusterOld.RpcRequest) {
	name := request.Rpcdata.Args[0].(string)
	logger.Info("node :" + name + " disconnected offline.")
	clusterserver.GlobalMaster.ChildDisconnected(name)
}