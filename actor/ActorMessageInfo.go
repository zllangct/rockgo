package Actor

import (
	"errors"
)

/*
	ActorMessage Information
*/
type ActorSession struct {
	Target        IActor
	Self          *ActorComponent
	ActorProxy    *ActorProxyComponent
	TargetActorID ActorID
	SelfActorID   ActorID
}

type ActorMessageInfo struct {
	Session *ActorSession
	Message IActorMessage
}

func (this *ActorMessageInfo) Emit(message IActorMessage) error{
	// Checking the message package, title and recipients are required
	if message.GetTitle() == "" {
		err:=errors.New("checking the message package title is required")
		return err
	}
	if this.Session.Target==nil && len(this.Session.TargetActorID)!=5  {
		err:=errors.New("target address can not be empty")
		return err
	}
	messageInfo := &ActorMessageInfo{
		Message:    message,
		Session:    this.Session,
	}
	//存在本地引用时直接对话
	if this.Session.Target != nil {
		messageInfo.Session=&ActorSession{
			Target:this.Session.Self,
			TargetActorID:this.Session.Self.ActorID,
		}
		err:= this.Session.Target.Tell(messageInfo)
		if err!=nil{
			//解除引用，防止内存泄漏,继续通过其他途径传递消息
			this.Session.Target=nil
		}else{
			return nil
		}
	}
	//不存在本地引用时通过ActorProxy对话
	if messageInfo.Session != nil {
		//TODO 走代理层
	}
	return nil
}
