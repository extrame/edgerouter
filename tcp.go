package edgerouter

import (
	"io"
	"net"
)

type TcpHandler interface {
	PacketReceived(bts []byte, conn *net.TCPConn) int
	Close(*net.TCPConn)
}

func handleTcpConn(conn *net.TCPConn, handler TcpHandler) {
	for {
		data := make([]byte, 512)
		read_length, err := conn.Read(data)
		if err != nil { // EOF, or worse
			if err == io.EOF {
				handler.Close(conn)
			}
			return
		}
		if read_length > 0 {
			handler.PacketReceived(data[0:read_length], conn)
		}
	}
}
