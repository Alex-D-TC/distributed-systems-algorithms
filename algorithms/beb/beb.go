package beb

import (
	"io/ioutil"
	"log"
	"net"
	"os"
	"sync"

	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/pfd"
	"github.com/alex-d-tc/distributed-systems-algorithms/protocol"
	"github.com/golang/protobuf/proto"

	"github.com/alex-d-tc/distributed-systems-algorithms/util"
)

type BestEffortBroadcastMessage struct {
	SourceHost    string
	CorrelationID uint64
	Message       []byte
}

type BestEffortBroadcast struct {
	hosts []string

	// BEB Main Functionality
	bebServicePort uint16

	// Outgoing event handlers
	deliverManager *onDeliverManager

	onHostDownListener <-chan string

	// Logging
	logger *log.Logger

	stateLock *sync.Mutex
}

func NewBestEffortBroadcast(bebServicePort uint16, hosts []string, pfd *pfd.PerfectFailureDetector) *BestEffortBroadcast {

	beb := &BestEffortBroadcast{
		hosts:          hosts,
		bebServicePort: bebServicePort,
		logger:         log.New(os.Stdout, "[BEB]", log.Ldate|log.Ltime),
		deliverManager: newOnDeliverManager(),
		stateLock:      &sync.Mutex{},
	}

	beb.onHostDownListener = pfd.AddOnProcessCrashedListener()

	// Run the server routine
	go beb.handleHostDown()
	go util.Listen(beb.bebServicePort, beb.handleConn, beb.logger)

	return beb
}

func (beb *BestEffortBroadcast) AddOnDeliverListener() <-chan BestEffortBroadcastMessage {
	return beb.deliverManager.AddListener()
}

func (beb *BestEffortBroadcast) Broadcast(correlationID uint64, message []byte) {

	beb.logger.Println("Sending broadcast message ", message)

	beb.stateLock.Lock()
	defer beb.stateLock.Unlock()

	bebMessage := &protocol.BEBMessage{
		CorrelationID: correlationID,
		Payload:       message,
	}

	rawBebMessage, err := proto.Marshal(bebMessage)
	if err != nil {
		beb.logger.Println(err)
	}

	for _, host := range beb.hosts {
		err := util.SendMessage(host, beb.bebServicePort, rawBebMessage)
		if err != nil {
			beb.logger.Println(err.Error())
		}
	}
}

func (beb *BestEffortBroadcast) handleHostDown() {
	for {
		downHost := <-beb.onHostDownListener

		beb.stateLock.Lock()
		for i, host := range beb.hosts {
			if host == downHost {
				beb.hosts[i] = beb.hosts[len(beb.hosts)-1]
				beb.hosts = beb.hosts[:len(beb.hosts)-1]
				break
			}
		}

		beb.stateLock.Unlock()
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

	bebMessage := &protocol.BEBMessage{}

	err = proto.Unmarshal(rawData, bebMessage)
	if err != nil {
		beb.logger.Println(err.Error())
		return
	}

	addr := conn.RemoteAddr().(*net.TCPAddr)

	beb.handleMessage(addr.IP.String(), bebMessage)
}

func (beb *BestEffortBroadcast) handleMessage(sourceHost string, bebMessage *protocol.BEBMessage) {

	msg := BestEffortBroadcastMessage{
		SourceHost:    sourceHost,
		CorrelationID: bebMessage.GetCorrelationID(),
		Message:       bebMessage.GetPayload(),
	}

	beb.deliverManager.Submit(msg)
}
