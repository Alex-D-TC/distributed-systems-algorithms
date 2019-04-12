package network

import (
	"fmt"
	"log"
	"net"
)

func listen(port int16, handler func(net.Conn), logger *log.Logger) {

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Printf("Listen init error: %s", err.Error())
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Print(err.Error())
		}

		go handler(conn)
	}
}

func sendWithGuarantee(conn net.Conn, message []byte) error {

	written := 0

	for written < len(message) {
		wrote, err := conn.Write(message[written:])
		if err != nil {
			return err
		}

		written += wrote
	}

	return nil
}
