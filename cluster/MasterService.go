package Cluster

import (
	"errors"
)

type NodeInfo struct {
	Time    int64
	Address string
	Role    []string
	AppName string
	Info    map[string]float32
}

type InquiryReply struct {
	Node	string
	Info    map[string]float32
}

type MasterService struct{
	master *MasterComponent
}

func (this *MasterService)init(master *MasterComponent) {
	this.master = master
}

func (this *MasterService) ReportNodeClose(addr string, reply *bool) error {
	this.master.locker.Lock()
	this.master.NodeClose(addr)
	this.master.locker.Unlock()
	*reply = true
	return nil
}

func (this *MasterService) ReportNodeInfo(args *NodeInfo, reply *bool) error {
	this.master.UpdateNodeInfo(args)
	*reply = true
	return nil
}

func (this *MasterService) NodeInquiry(args []string, reply *[]*InquiryReply) error {
	//logger.Debug("Inquiry :",args)
	res,err:= this.master.NodeInquiry(args,false)
	*reply =res
	return err
}

func (this *MasterService) NodeInquiryDetail(args []string, reply *[]*InquiryReply) error {
	res,err:= this.master.NodeInquiry(args,true)
	*reply =res
	return err
}

type NodeInfoSyncReply struct {
	Nodes   map[string]*NodeInfo
	NodeLog *NodeLogs
}

func (this *MasterService) NodeInfoSync(args string, reply *NodeInfoSyncReply) error {
	if args != "sync" {
		return errors.New("call service [ NodeInfoSynchronous ],has wrong argument")
	}
	*reply=NodeInfoSyncReply{
		Nodes:   this.master.NodesCopy(),
		NodeLog: this.master.NodesLogsCopy(),
	}
	return nil
}

type NodeLog struct {
	Time int64
	Log string
	Type int
}

type NodeLogs struct {
	BufferSize int
	Logs []*NodeLog
}

func (this *NodeLogs)Add(log *NodeLog)  {
	if len(this.Logs)< this.BufferSize{
		this.Logs= append(this.Logs, log)
	}else{
		this.Logs= append(this.Logs[1:],log)
	}
}

func (this *NodeLogs) Get(time int64) []*NodeLog {
	for key, value := range this.Logs {
		if value.Time > time {
			return this.Logs[key:]
		}
	}
	return []*NodeLog{}
}

