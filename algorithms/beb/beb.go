package beb

import (
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/alex-d-tc/distributed-systems-algorithms/protocol"
	"github.com/golang/protobuf/proto"

	"github.com/alex-d-tc/distributed-systems-algorithms/util"
)

type BestEffortBroadcastMessage struct {
	SourceHost string
	Message    []byte
}

type BestEffortBroadcast struct {
	hosts []string

	// BEB Main Functionality
	bebServicePort uint16

	// Outgoing event handlers
	deliverManager *onDeliverManager

	// Logging
	logger *log.Logger
}

func NewBestEffortBroadcast(bebServicePort uint16, hosts []string) *BestEffortBroadcast {

	beb := &BestEffortBroadcast{
		hosts:          hosts,
		bebServicePort: bebServicePort,
		logger:         log.New(os.Stdout, "[BEB]", log.Ldate|log.Ltime),
		deliverManager: newOnDeliverManager(),
	}

	// Run the server routine
	go util.Listen(beb.bebServicePort, beb.handleConn, beb.logger)

	return beb
}

func (beb *BestEffortBroadcast) AddOnDeliverListener() <-chan BestEffortBroadcastMessage {
	return beb.deliverManager.AddListener()
}

func (beb *BestEffortBroadcast) Broadcast(message []byte) {
	
	beb.logger.Println("Sending broadcast message ", message)

	bebMessage := &protocol.BEBMessage{
		Payload: message,
	}

	rawBebMessage, err := proto.Marshal(bebMessage)
	if err != nil {
		beb.logger.Println(err)
	}

	for _, host := range beb.hosts {
		err, _ := util.SendMessage(host, beb.bebServicePort, rawBebMessage)
		if err != nil {
			beb.logger.Println(err.Error())
		}
	}
}

func (beb *BestEffortBroadcast) handleConn(conn net.Conn) {
	defer conn.Close()

	beb.logger.Println("Received BEB message from ", conn.RemoteAddr().String())

	rawData, err := ioutil.ReadAll(conn)
	if err != nil {
		beb.logger.Println(err.Error())
		return
	}

	err, wireMessage := util.GetWireMessage(rawData)
	if err != nil {
		beb.logger.Printf("Error when decoding wire message: %s\n", err.Error())
		return
	}

	bebMessage := &protocol.BEBMessage{}

	err = proto.Unmarshal(wireMessage.GetPayload(), bebMessage)
	if err != nil {
		beb.logger.Println(err.Error())
		return
	}

	addr := conn.RemoteAddr().(*net.TCPAddr)

	beb.handleMessage(addr.String(), bebMessage)
}

func (beb *BestEffortBroadcast) handleMessage(sourceHost string, bebMessage *protocol.BEBMessage) {

	msg := BestEffortBroadcastMessage{
		SourceHost: sourceHost,
		Message:    bebMessage.GetPayload(),
	}

	beb.deliverManager.Submit(msg)
}
