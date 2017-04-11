package edgerouter

import (
	"io"
	"net"

	"github.com/golang/glog"
)

type TcpHandler interface {
	PacketReceived(bts []byte, conn *net.TCPConn) int
	Close(*net.TCPConn)
}

func handleTcpConn(conn *net.TCPConn, handler TcpHandler) {
	for {
		data := make([]byte, 512)
		readLength, err := conn.Read(data)
		glog.Infof("got %d bytes", readLength)
		if err != nil { // EOF, or worse
			if err == io.EOF {
				handler.Close(conn)
			}
			return
		}
		if readLength > 0 {
			handler.PacketReceived(data[0:readLength], conn)
		}
	}
}
