package Actor

/*
	ActorMessage Information
*/

type ActorMessageInfo struct {
	Sender IActor
	Message *ActorMessage
}

type ActorRpcMessageInfo struct {
	Target  ActorID
	Sender ActorID
	Message *ActorMessage
}

