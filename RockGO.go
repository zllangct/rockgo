package RockGO

import (
	_ "github.com/zllangct/RockGO/Network"
	_ "github.com/zllangct/RockGO/timer"
	"github.com/zllangct/RockGO/telnetcmd"
	"github.com/zllangct/RockGO/clusterserver"
	"github.com/zllangct/RockGO/sys_rpc"
	"github.com/zllangct/RockGO/utils"
	"github.com/zllangct/RockGO/fserver"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/logger"
	"fmt"
	"github.com/zllangct/RockGO/RockInterface"
)

func NewXingoTcpServer() RockInterface.Iserver{
	//do something
	//debugport 是否开放
	if utils.GlobalObject.DebugPort > 0{
		if utils.GlobalObject.Host != ""{
			fserver.NewTcpServer("telnet_server", "tcp4", utils.GlobalObject.Host,
				utils.GlobalObject.DebugPort, 100, cluster.NewTelnetProtocol()).Start()
		}else{
			fserver.NewTcpServer("telnet_server", "tcp4", "127.0.0.1",
				utils.GlobalObject.DebugPort, 100, cluster.NewTelnetProtocol()).Start()
		}
		logger.Debug(fmt.Sprintf("telnet tool start: %s:%d.", utils.GlobalObject.Host, utils.GlobalObject.DebugPort))

	}

	//add command
	if utils.GlobalObject.CmdInterpreter != nil{
		utils.GlobalObject.CmdInterpreter.AddCommand(telnetcmd.NewPprofCpuCommand())
	}

	s := fserver.NewServer()
	return s
}

func NewXingoMater(cfg string) *clusterserver.Master{
	s := clusterserver.NewMaster(cfg)
	//add rpc
	s.AddRpcRouter(&sys_rpc.MasterRpc{})
	//add command
	if utils.GlobalObject.CmdInterpreter != nil{
		utils.GlobalObject.CmdInterpreter.AddCommand(telnetcmd.NewPprofCpuCommand())
		utils.GlobalObject.CmdInterpreter.AddCommand(telnetcmd.NewCloseServerCommand())
		utils.GlobalObject.CmdInterpreter.AddCommand(telnetcmd.NewReloadCfgCommand())
	}
	return s
}

func NewXingoCluterServer(nodename, cfg string) *clusterserver.ClusterServer{
	s := clusterserver.NewClusterServer(nodename,cfg)
	//add rpc
	s.AddRpcRouter(&sys_rpc.ChildRpc{})
	s.AddRpcRouter(&sys_rpc.RootRpc{})
	//add cmd
	if utils.GlobalObject.CmdInterpreter != nil{
		utils.GlobalObject.CmdInterpreter.AddCommand(telnetcmd.NewPprofCpuCommand())
	}
	return s
}