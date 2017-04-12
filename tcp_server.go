package edgerouter

import (
	"errors"
	"fmt"
	"net"

	"github.com/golang/glog"
)

type TcpServer struct {
	ctl   Controller
	Port  int
	conns map[string]*net.TCPConn
}

func (s *TcpServer) Init() (err error) {
	s.conns = make(map[string]*net.TCPConn)

	return err
}

func (c *TcpServer) PacketReceived(bts []byte, conn *net.TCPConn) int {
	return c.ctl.OnReceived(bts, conn)
}

func (c *TcpServer) SetController(ctrl Controller) {
	c.ctl = ctrl
}

func (c *TcpServer) Connect(to string) error {
	if addr, err := net.ResolveIPAddr("ip", to); err == nil {
		if _, ok := c.conns[addr.String()]; !ok {
			return errors.New("no such tcp client for " + addr.String())
		}
		return nil
	} else {
		return err
	}
}

func (c *TcpServer) GetConn(to string, unexceptedConns []net.Conn) (conn net.Conn, err error) {
	var addr *net.IPAddr
	if to == "any" {
	OUTER:
		for _, conn := range c.conns {
			for _, unexcepted := range unexceptedConns {
				if conn == unexcepted {
					continue OUTER
				}
			}
			return conn, nil
		}
		return nil, NoSuchConnection
	}
	if addr, err = net.ResolveIPAddr("ip", to); err == nil {
		var ok bool
		if conn, ok = c.conns[addr.String()]; ok {
		} else {
			err = NoSuchConnection
		}
	}
	return
}

func (c *TcpServer) DeleteConn(toDel net.Conn) {
	toDel.Close()
	for key, conn := range c.conns {
		if conn == toDel {
			delete(c.conns, key)
		}
	}
}

func (c *TcpServer) Close(toDel *net.TCPConn) {
	c.DeleteConn(toDel)
}

// func (c *TcpServer) Send(msg *BytesMessage) (err error) {
// 	var addr *net.IPAddr
// 	if addr, err = net.ResolveIPAddr("ip", msg.To); err == nil {
// 		if conn, ok := c.conns[addr.IP.String()]; ok {
// 			_, err = conn.Write(msg.Message)
// 		} else {
// 			err = errors.New("no such connection for " + addr.IP.String())
// 		}
// 	}

// 	return err
// }

func (c *TcpServer) String() string {
	return fmt.Sprintf("tcp server(%p) listened on (:%d)", c, c.Port)
}

func (s *TcpServer) Run() {
	var addr *net.TCPAddr
	var err error
	var listener *net.TCPListener
	if addr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", s.Port)); err == nil {
		listener, err = net.ListenTCP("tcp", addr)
	}
	for {
		if conn, err := listener.AcceptTCP(); err == nil {
			addr := conn.RemoteAddr().(*net.TCPAddr)
			glog.Infof("got connection from (%s)", addr)
			s.conns[addr.IP.String()] = conn
			go handleTcpConn(conn, s)
		}
	}
}
