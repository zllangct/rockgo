package RockInterface

type IWriter interface {
	Send([]byte) error
	GetProperty(string) (interface{}, error)
	SetProperty(string, interface{})
	RemoveProperty(string)
	//GetConnection() *net.TCPConn
}
