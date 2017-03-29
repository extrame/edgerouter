package edgerouter

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
)

type TcpServer struct {
	Port     int
	listener *net.TCPListener
}

type TcpHandler interface {
	PacketReceived(bts []byte, conn *net.TCPConn) int
}

func (u *TcpServer) Run(ctx context.Context, handler interface{}) (context.Context, error) {
	fmt.Println("run tcp server")
	if uh, ok := handler.(TcpHandler); ok {
		laddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(u.Port))
		if err == nil {
			fmt.Println("listening on " + strconv.Itoa(u.Port))
			u.listener, err = net.ListenTCP("tcp", laddr)
			ctx := context.WithValue(ctx, "tcp-listener", u.listener)
			if err != nil {
				return nil, err
			}
			go u.listenTcp(ctx, uh)
			return ctx, nil
		}
		return nil, err
	}
	return nil, errors.New("the plugin is not a Tcp handler with DatagramReceived function")
}

func (u *TcpServer) listenTcp(ctx context.Context, handler TcpHandler) {
	for {
		if conn, err := u.listener.AcceptTCP(); err != nil {
			go handleTcpConn(conn, handler)
		} else {
			fmt.Println(err)
		}
	}
}

func handleTcpConn(conn *net.TCPConn, handler TcpHandler) {
	for {
		data := make([]byte, 512)
		read_length, err := conn.Read(data)
		if err != nil { // EOF, or worse
			fmt.Println(err)
			return
		}
		if read_length > 0 {
			go panicWrapping(func() {
				handler.PacketReceived(data[0:read_length], conn)
			})
		}
	}
}

// func (u *TcpServer) DatagramReceived(bts []byte, addr *net.TcpAddr) int {
// 	panic("you should overwrite DatagramReceived function for Tcp server")
// 	return len(bts)
// }
