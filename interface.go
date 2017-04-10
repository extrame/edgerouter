package edgerouter

import (
	"net"
)

type Transport interface {
	Connect(string) error
	// Send(*BytesMessage) error
	SetController(Controller)
	GetConn(addr string) (net.Conn, error)
	DeleteConn(net.Conn)
}

type Server interface {
	Run()
}
