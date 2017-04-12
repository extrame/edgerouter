package edgerouter

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/golang/glog"
)

type SerialSeeker struct {
	Period          string
	TimeOut         string
	trans           Transport
	waitedDevice    Device
	chanAvailable   chan bool
	chanFinish      chan bool
	to              time.Duration
	duration        time.Duration
	handler         SerialSeekHandler
	preferredConn   map[string]net.Conn
	unexceptedConns map[string][]net.Conn
}

type SerialSeekHandler interface {
	PacketSend() []*BytesMessage
	SeekReceived([]byte, Device) (handled_length int, shouldStartNew bool)
}

func (s *SerialSeeker) Init() (err error) {
	glog.Infoln("init serial seeker")
	s.chanFinish = make(chan bool)
	s.chanAvailable = make(chan bool, 1)
	s.chanAvailable <- true
	s.preferredConn = make(map[string]net.Conn)
	s.unexceptedConns = make(map[string][]net.Conn)
	if s.to, err = time.ParseDuration(s.TimeOut); err == nil {
		s.duration, err = time.ParseDuration(s.Period)
	}
	return err
}

func (s *SerialSeeker) SetHandler(handler interface{}) error {
	if uh, ok := handler.(SerialSeekHandler); ok {
		s.handler = uh
		return nil
	}
	return errors.New("the plugin is not a serial seek handler with PacketSend and SeekReceived function")
}

func (s *SerialSeeker) SetTransport(t Transport) {
	s.trans = t
}

func (s *SerialSeeker) Run() {
	for {
		select {
		case <-time.After(s.duration):
			msgs := s.handler.PacketSend()
			glog.Infof("periodly seeking for %d messages", len(msgs))
			for _, msg := range msgs {
				go s.seek(msg)
			}
		}
	}
}

func (s *SerialSeeker) seek(msg *BytesMessage) (err error) {
	var conn net.Conn
	var isPreferred bool
	if msg.To != "any" {
		if err = s.trans.Connect(msg.To); err != nil {
			goto errHandling
		}
	} else {
		conn, isPreferred = s.preferredConn[msg.For.DeviceID()]
	}
	glog.Info("wait for available chan")
	<-s.chanAvailable
	glog.Info("send for", msg.For)
	s.waitedDevice = msg.For
	if conn == nil {
		if conn, err = s.trans.GetConn(msg.To, s.unexceptedConns[msg.For.DeviceID()]); err != nil {
			s.chanAvailable <- false
			s.unexceptedConns[msg.For.DeviceID()] = s.unexceptedConns[msg.For.DeviceID()][:0]
			goto errHandling
		}
	}
	glog.Info("in conn", conn)
	if _, err = conn.Write(msg.Message); err != nil {
		s.trans.DeleteConn(conn)
		if isPreferred {
			delete(s.preferredConn, msg.For.DeviceID())
		}
		goto errHandling
	} else {
		glog.Infoln("...ok!")
	}
	select {
	case <-time.After(s.to):
		glog.Infoln("timeout")
		if isPreferred {
			delete(s.preferredConn, msg.For.DeviceID())
		}
		s.unexceptedConns[msg.For.DeviceID()] = append(s.unexceptedConns[msg.For.DeviceID()], conn)
		s.chanAvailable <- false
	case <-s.chanFinish:
		if !isPreferred {
			s.preferredConn[msg.For.DeviceID()] = conn
		}
		glog.Infoln("finished")
		s.chanAvailable <- true
	}
	return nil
errHandling:
	glog.Errorln(err)
	return err
}

func (s *SerialSeeker) OnReceived(bts []byte, conn net.Conn) int {
	_, shouldStartNew := s.handler.SeekReceived(bts, s.waitedDevice)
	if shouldStartNew {
		s.chanFinish <- true
	}
	return len(bts)
}

func (s *SerialSeeker) String() string {
	return fmt.Sprintf("serial seeking(%p)", s)
}
