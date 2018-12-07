package Cluster

var location *LocationComponent

type LocationService struct{}

func (this *LocationService)init(mlocation *LocationComponent) {
	location = mlocation
}

func (this *LocationService) NodeInquiry(args *string, reply *[]*InquiryReply) error {
	res,err:= location.NodeInquiry(*args)
	reply =&res
	return err
}