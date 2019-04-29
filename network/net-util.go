package network

import (
	"fmt"
	"log"
	"net"
)

func SendMessage(hostname string, port int16, message []byte) error {

	// Initiate the connection
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
	if err != nil {
		return err
	}
	defer conn.Close()

	// Send the data
	return SendWithGuarantee(conn, message)
}

func Listen(port int16, handler func(net.Conn), logger *log.Logger) {

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

func SendWithGuarantee(conn net.Conn, message []byte) error {

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
