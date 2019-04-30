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

type BestEffortBroadcast struct {
	hosts []string

	// BEB Main Functionality
	bebServicePort      uint16
	bebDeliverListeners []chan<- []byte

	// Logging
	logger *log.Logger
}

func NewBestEffortBroadcast(hosts []string, bebServicePort uint16) *BestEffortBroadcast {

	beb := &BestEffortBroadcast{
		hosts:               hosts,
		bebServicePort:      bebServicePort,
		bebDeliverListeners: []chan<- []byte{},
		logger:              log.New(os.Stdout, "[BEB]", log.Ldate|log.Ltime),
	}

	// Run the server routine
	go util.Listen(beb.bebServicePort, beb.handleConn, beb.logger)

	return beb
}

func (beb *BestEffortBroadcast) RegisterOnDeliverHandler(listener chan<- []byte) {
	beb.bebDeliverListeners = append(beb.bebDeliverListeners, listener)
}

func (beb *BestEffortBroadcast) Broadcast(message []byte) {
	for _, host := range beb.hosts {
		err := util.SendMessage(host, beb.bebServicePort, message)
		if err != nil {
			beb.logger.Println(err.Error())
		}
	}
}

func (beb *BestEffortBroadcast) handleConn(conn net.Conn) {
	defer conn.Close()

	rawData, err := ioutil.ReadAll(conn)
	if err != nil {
		beb.logger.Println(err.Error())
		return
	}

	bebMessage := &protocol.BEBMessage{}

	err = proto.Unmarshal(rawData, bebMessage)
	if err != nil {
		beb.logger.Println(err.Error())
		return
	}

	beb.handleMessage(bebMessage)
}

func (beb *BestEffortBroadcast) handleMessage(bebMessage *protocol.BEBMessage) {
	for _, listener := range beb.bebDeliverListeners {
		listener <- bebMessage.Payload
	}
}
