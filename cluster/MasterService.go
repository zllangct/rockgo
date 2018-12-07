package Cluster

import "errors"

var master *MasterComponent

type NodeInfo struct {
	Address      string
	Group   []string
	AppName []string
	Info    map[string]float32
}

type InquiryReply struct {
	Node	string
	Info    map[string]float32
}

type MasterService struct{}

func (this *MasterService)init(mmaster *MasterComponent) {
	master = mmaster
}

func (this *MasterService) ReportNodeInfo(args *NodeInfo, reply *bool) error {
	master.UpdateNodeInfo(args)
	*reply = true
	return nil
}

func (this *MasterService) NodeInquiry(args *string, reply *[]*InquiryReply) error {
	res,err:= master.NodeInquiry(*args)
	reply =&res
	return err
}

func (this *MasterService) NodeInfoSynchronous(args string, reply *map[string]*NodeInfo) error {
	if args != "sync" {
		return errors.New("call service [ NodeInfoSynchronous ],has wrong argument")
	}
	nodes := master.NodesCopy()
	reply=&nodes
	return nil
}