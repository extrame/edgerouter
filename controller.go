package edgerouter

import (
	"net"
)

type Controller interface {
	OnReceived(bts []byte, conn net.Conn) int
	SetTransport(Transport)
	SetHandler(handler interface{}) error
	Run()
}
