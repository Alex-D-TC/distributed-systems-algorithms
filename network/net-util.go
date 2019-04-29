package network

import (
	"fmt"
	"log"
	"net"
)

// SendMessage sends a message to the hostname:port remote address
// The message is guaranteed to be fully delivered, or fail with an error
func SendMessage(hostname string, port uint16, message []byte) error {

	// Initiate the connection
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
	if err != nil {
		return err
	}
	defer conn.Close()

	// Send the data
	return SendWithGuarantee(conn, message)
}

// Listen creates a listener on all interfaces on port port, with a custom connection handler
// The logger is used to log listen failures
func Listen(port uint16, handler func(net.Conn), logger *log.Logger) {

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

// SendWithGuarantee sends a message on the connection conn
// The message is guaranteed to be fully delivered, or the function returns with an error
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
