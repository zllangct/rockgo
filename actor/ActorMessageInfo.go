package Actor

/*
	ActorMessage Information
*/

type ActorMessageInfo struct {
	Sender IActor
	Message *ActorMessage
	Reply *ActorMessage
}

type ActorRpcMessageInfo struct {
	Target  ActorID
	Sender ActorID
	Message *ActorMessage
	Reply *ActorMessage
}

