package edgerouter

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/golang/glog"
)

type SerialSeeker struct {
	Period        string
	TimeOut       string
	trans         Transport
	waitedDevice  Device
	chanAvailable chan bool
	chanFinish    chan bool
	to            time.Duration
	duration      time.Duration
	handler       SerialSeekHandler
}

type SerialSeekHandler interface {
	PacketSend() []*BytesMessage
	SeekReceived([]byte, Device) (handled_length int, shouldStartNew bool)
}

func (s *SerialSeeker) Init() (err error) {
	s.chanFinish = make(chan bool)
	s.chanAvailable = make(chan bool, 1)
	s.chanAvailable <- true
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
			glog.Infoln("periodly seeking")
			msgs := s.handler.PacketSend()
			for _, msg := range msgs {
				go s.seek(msg)
			}
		}
	}
}

func (s *SerialSeeker) seek(msg *BytesMessage) (err error) {
	var conn net.Conn
	if err = s.trans.Connect(msg.To); err != nil {
		goto errHandling
	}
	<-s.chanAvailable
	s.waitedDevice = msg.For
	glog.Info("start for", msg.For)
	if conn, err = s.trans.GetConn(msg.To); err != nil {
		goto errHandling
	}
	if _, err = conn.Write(msg.Message); err != nil {
		s.trans.DeleteConn(conn)
		goto errHandling
	}
	select {
	case <-time.After(s.to):
		glog.Infoln("timeout")
		s.chanAvailable <- false
	case <-s.chanFinish:
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
