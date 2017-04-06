package edgerouter

import (
	"fmt"
	"testing"
)

import "time"

type TestSeeker struct {
	SerialSeeker
	TcpServer
}

func (t *TestSeeker) PacketSend() []*BytesMessage {
	fmt.Println("....")
	return []*BytesMessage{}
}
func (t *TestSeeker) SeekReceived([]byte, Device) (handled_length int, shouldStartNew bool) {
	return 0, false
}

func TestInit(t *testing.T) {
	er := Organize("test", TestSeeker{})
	er.ConfigByString(`
	[edgerouter]
url = "127.0.0.1:9099"
dburl = "10.11.22.123"
dbuser = "rongshu"
dbpassword = "MinkTech2501"
dbname = "weifang"
dbport = 12306
period = "3s"
uselistenedaddr = true
timeout = "3s"
port = 6100
[[edgerouter.devices."10.11.22.35:4196"]]
addr = "712910"
type = "SZY"
[[edgerouter.devices."10.11.22.35:4196"]]
addr = "713205"
type = "SZY"
[[edgerouter.devices."10.11.22.35:4196"]]
addr = "713326"
type = "WYJ"
`)
	go er.Run()
	select {
	case <-time.After(time.Second * 20):
		return
	}
}
