package edgerouter

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

type ConcurrentTcpSeeker struct {
	Period string
	Port   int
}

type ConcurrentTcpSeekHandler interface {
	PacketSend() []*BytesMessage
	PacketReceived(bts []byte, conn *net.TCPConn) int
}

func (u *ConcurrentTcpSeeker) Run(ctx context.Context, handler interface{}) (context.Context, error) {
	var err error
	var d time.Duration
	if d, err = time.ParseDuration(u.Period); err != nil {
		return ctx, err
	} else {
		fmt.Printf("tcp send package by (%s) period\n", d)
	}
	if uh, ok := handler.(ConcurrentTcpSeekHandler); ok {
		go u.handleTcpSeek(ctx, d, uh)
		return ctx, err
	}
	return ctx, errors.New("the plugin is not a udp seek handler with DatagramSend function")
}

func (u *ConcurrentTcpSeeker) handleTcpSeek(ctx context.Context, d time.Duration, handler ConcurrentTcpSeekHandler) {
	var conns = make(map[string]*net.TCPConn)
	for {
		select {
		case <-time.After(d):
			fmt.Println(".")
			msgs := handler.PacketSend()
			for _, msg := range msgs {
				var err error
				var addr *net.TCPAddr
				var conn *net.TCPConn
				if addr, err = net.ResolveTCPAddr("tcp", msg.To); err == nil {
					var ok bool
					if conn, ok = conns[addr.String()]; !ok {
						var localPort *net.TCPAddr
						if u.Port != 0 {
							localPort, _ = net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", u.Port))
						}
						conn, err = net.DialTCP("tcp", localPort, addr)
						go handleTcpConn(conn, handler)
						conns[msg.To] = conn
					}
					_, err = conn.Write(msg.Message)
				}
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}
