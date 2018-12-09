package Cluster

import "errors"

type NodeInfo struct {
	Address      string
	Group   []string
	AppName  string
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

func (this *MasterService) ReportNodeInfo(args *NodeInfo, reply *bool) error {
	this.master.UpdateNodeInfo(args)
	*reply = true
	return nil
}

func (this *MasterService) NodeInquiry(args *string, reply *[]*InquiryReply) error {
	res,err:= this.master.NodeInquiry(*args,false)
	reply =&res
	return err
}

func (this *MasterService) NodeInquiryDetail(args *string, reply *[]*InquiryReply) error {
	res,err:= this.master.NodeInquiry(*args,true)
	reply =&res
	return err
}

func (this *MasterService) NodeInfoSync(args string, reply *map[string]*NodeInfo) error {
	if args != "sync" {
		return errors.New("call service [ NodeInfoSynchronous ],has wrong argument")
	}
	nodes := this.master.NodesCopy()
	reply=&nodes
	return nil
}