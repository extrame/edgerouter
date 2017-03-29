package edgerouter

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
)

type UdpServer struct {
	Port int
	conn *net.UDPConn
}

type UdpHandler interface {
	DatagramReceived(bts []byte, addr *net.UDPAddr) int
}

func (u *UdpServer) Run(ctx context.Context, handler interface{}) (context.Context, error) {
	if uh, ok := handler.(UdpHandler); ok {
		laddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(u.Port))
		if err == nil {
			fmt.Println("listening on " + strconv.Itoa(u.Port))
			u.conn, err = net.ListenUDP("udp", laddr)
			ctx := context.WithValue(ctx, "listened-conn", u.conn)
			if err != nil {
				return nil, err
			}
			go u.handleUdpConnection(ctx, uh)
			return ctx, nil
		}
		return nil, err
	}
	return nil, errors.New("the plugin is not a udp handler with DatagramReceived function")
}

func (u *UdpServer) handleUdpConnection(ctx context.Context, handler UdpHandler) {
	for {
		data := make([]byte, 512)
		read_length, remoteAddr, err := u.conn.ReadFromUDP(data[0:])
		if err != nil { // EOF, or worse
			return
		}
		if read_length > 0 {
			go panicWrapping(func() {
				handledLength := handler.DatagramReceived(data[0:read_length], remoteAddr)
				data = data[handledLength:]
			})
		}
	}
}

// func (u *UdpServer) DatagramReceived(bts []byte, addr *net.UDPAddr) int {
// 	panic("you should overwrite DatagramReceived function for udp server")
// 	return len(bts)
// }
