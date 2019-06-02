package util

import (
	"fmt"
	"log"
	"net"
	"math/rand"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/alex-d-tc/distributed-systems-algorithms/protocol"
)

// SendMessage sends a message to the hostname:port remote address
// The message is guaranteed to be fully delivered, or fail with an error
func SendMessage(hostname string, port uint16, message []byte) (error, uint64) {

	// Initiate the connection
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
	if err != nil {
		return err, 0
	}
	defer conn.Close()

	// rand.Int63n returns a positive integer anyway
	nonce := uint64(rand.Int63n(time.Now().Unix()))

	// Wrap the data in a wire message
	wireMsg := &protocol.WireMessage{
		Nonce: nonce,
		Payload: message,
	}

	rawMessage, err := proto.Marshal(wireMsg)
	if err != nil {
		return err, 0
	}

	// Send the data
	err = WriteWithGuarantee(conn, rawMessage)
	if err != nil {
		return err, 0
	}
	return nil, nonce
}

// Listen creates a listener on all interfaces on port port, with a custom connection handler
// The logger is used to log listen failures
func Listen(port uint16, handler func(net.Conn), logger *log.Logger) {

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
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

func GetWireMessage(rawMessage []byte) (error, *protocol.WireMessage) {
	wireM := &protocol.WireMessage{}

	err := proto.Unmarshal(rawMessage, wireM)
	if err != nil {
		return err, nil
	}

	return nil, wireM
}
