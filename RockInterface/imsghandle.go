package RockInterface

type Imsghandle interface {
	DeliverToMsgQueue(interface{})
	DoMsgFromGoRoutine(interface{})
	AddRouter(interface{})
	StartWorkerLoop(int)
}
