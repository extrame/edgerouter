package edgerouter

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/golang/glog"
)

type SerialTcpSeeker struct {
	Period  string
	Port    int
	TimeOut string
}

type Device interface {
	DeviceID() string
	DeviceType() string
}

type SerialTcpSeekHandler interface {
	PacketSend() []*BytesMessage
	SeekReceived([]byte, Device) (handled_length int, shouldStartNew bool)
}

func (u *SerialTcpSeeker) Run(ctx context.Context, handler interface{}) (context.Context, error) {
	var err error
	var d, to time.Duration
	if d, err = time.ParseDuration(u.Period); err != nil {
		return ctx, err
	} else {
		fmt.Printf("tcp send package by (%s) period\n", d)
	}
	if to, err = time.ParseDuration(u.TimeOut); err != nil {
		return ctx, err
	}
	if uh, ok := handler.(SerialTcpSeekHandler); ok {
		go u.handleTcpSeek(ctx, d, to, uh)
		return ctx, err
	}
	return ctx, errors.New("the plugin is not a tcp serial seek handler with PacketSend and SeekReceived function")
}

type Seeking struct {
	conn       *net.TCPConn
	cha        chan bool
	device     Device
	to         time.Duration
	handler    SerialTcpSeekHandler
	toReminder *time.Timer
}

func (s *Seeking) PacketReceived(bts []byte, conn *net.TCPConn) int {
	_, shouldStartNew := s.handler.SeekReceived(bts, s.device)
	if shouldStartNew {
		s.cha <- true
	}
	return len(bts)
}

func (u *SerialTcpSeeker) handleTcpSeek(ctx context.Context, d, to time.Duration, handler SerialTcpSeekHandler) {
	var seekings = make(map[string]*Seeking)
	for {
		select {
		case <-time.After(d):
			msgs := handler.PacketSend()
			for _, msg := range msgs {
				var err error
				var addr *net.TCPAddr
				var seeking *Seeking
				if addr, err = net.ResolveTCPAddr("tcp", msg.To); err == nil {
					var ok bool
					if seeking, ok = seekings[addr.String()]; !ok {
						var localPort *net.TCPAddr
						if u.Port != 0 {
							localPort, _ = net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", u.Port))
						}
						var conn *net.TCPConn
						conn, err = net.DialTCP("tcp", localPort, addr)
						seeking = &Seeking{
							conn:    conn,
							cha:     make(chan bool, 1),
							to:      to,
							handler: handler,
						}
						go handleTcpConn(conn, seeking)
						seekings[addr.String()] = seeking
						seeking.cha <- true
					}
					<-seeking.cha
					if seeking.toReminder != nil {
						seeking.toReminder.Stop()
					}
					seeking.device = msg.For
					glog.Info("start for", msg.For)
					_, err = seeking.conn.Write(msg.Message)
					seeking.toReminder = time.AfterFunc(seeking.to, func() {
						glog.Infoln("timeout")
						seeking.cha <- false
					})
				}
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}
