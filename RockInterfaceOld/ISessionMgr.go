package RockInterface

type ISessionMgr interface {
	Add(ISession)
	Remove(ISession) error
	Get(uint32) (ISession, error)
	Len() int
}
