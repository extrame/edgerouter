package edgerouter

import (
	"fmt"
	"net"
)

type TcpHandler interface {
	PacketReceived(bts []byte, conn *net.TCPConn) int
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
			handler.PacketReceived(data[0:read_length], conn)
		}
	}
}
