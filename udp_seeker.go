package edgerouter

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

type UdpSeeker struct {
	Period          string
	UseListenedAddr bool
	conn            *net.UDPConn
}

type UdpSeekHandler interface {
	DatagramSend() []*BytesMessage
}

func (u *UdpSeeker) Run(ctx context.Context, handler interface{}) (context.Context, error) {
	var err error
	var d time.Duration
	if d, err = time.ParseDuration(u.Period); err != nil {
		return ctx, err
	} else {
		fmt.Printf("udp send package by (%s) period\n", d)
	}
	if u.UseListenedAddr {
		conn := ctx.Value("listened-conn")
		if conn != nil {
			u.conn = conn.(*net.UDPConn)
		} else {
			log.Fatal("udp conn is nil")
		}
	}
	if uh, ok := handler.(UdpSeekHandler); ok {
		go u.handleUdpSeek(ctx, d, uh)
		return ctx, err
	}
	return ctx, errors.New("the plugin is not a udp seek handler with DatagramSend function")
}

func (u *UdpSeeker) handleUdpSeek(ctx context.Context, d time.Duration, handler UdpSeekHandler) {
	for {
		select {
		case <-time.After(d):
			fmt.Println(".")
			msgs := handler.DatagramSend()
			for _, msg := range msgs {
				var err error
				var addr *net.UDPAddr
				fmt.Println(msg)
				if addr, err = net.ResolveUDPAddr("udp", msg.To); err == nil {
					conn := u.conn
					if conn == nil && !u.UseListenedAddr {
						conn, err = net.DialUDP("udp", nil, addr)
					}
					if err == nil {
						_, err = conn.WriteToUDP(msg.Message, addr)
					}
				}
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}
