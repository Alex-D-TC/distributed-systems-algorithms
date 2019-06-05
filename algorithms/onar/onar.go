package onar

import (
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sync"

	"github.com/golang/protobuf/proto"

	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/beb"
	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/pfd"
	"github.com/alex-d-tc/distributed-systems-algorithms/protocol"
	"github.com/alex-d-tc/distributed-systems-algorithms/util"
	"github.com/alex-d-tc/distributed-systems-algorithms/util/environment"
)

var ackBytes = [...]byte{0, 4, 5, 1}

const correlationID = 1

type ONAR struct {

	// Algorithm-relevant data
	timestamp uint64
	val       int64

	correctProcesses []string
	writeset         map[string]bool

	readval int64
	reading bool

	// Algorithm dependencies
	beb      *beb.BestEffortBroadcast
	pfd      *pfd.PerfectFailureDetector
	onarPort uint16

	// Exposed events
	onReadReturn  *onReadReturnManager
	onWriteReturn *onWriteReturnManager

	// External event queues
	pfdOnCrashedListener <-chan string
	bebOnDeliverListener <-chan beb.BestEffortBroadcastMessage

	// Logging
	logger *log.Logger

	// Extras
	stateLock *sync.Mutex
}

func NewONAR(env *environment.NetworkEnvironment, beb *beb.BestEffortBroadcast, pfd *pfd.PerfectFailureDetector) *ONAR {

	onar := &ONAR{
		onarPort:         env.GetONARPort(),
		beb:              beb,
		pfd:              pfd,
		timestamp:        1,
		val:              0,
		correctProcesses: env.GetHosts(),
		writeset:         make(map[string]bool, 0),
		readval:          0,
		reading:          false,
		onReadReturn:     NewOnReadReturnManager(),
		onWriteReturn:    NewOnWriteReturnManager(),
		logger:           log.New(os.Stdout, "[ONAR]", log.Ldate|log.Ltime),
		stateLock:        &sync.Mutex{},
	}

	onar.pfdOnCrashedListener = pfd.AddOnProcessCrashedListener()
	onar.bebOnDeliverListener = beb.AddOnDeliverListener()

	go onar.handleBebDeliver()
	go onar.handleProcessCrash()
	go util.Listen(onar.onarPort, onar.handleAckMessage, onar.logger)

	return onar
}

func (onar *ONAR) AddOnWriteReturnListener() <-chan WriteReturn {
	return onar.onWriteReturn.AddListener()
}

func (onar *ONAR) AddOnReadReturnListener() <-chan ReadReturn {
	return onar.onReadReturn.AddListener()
}

func (onar *ONAR) Read() error {

	onar.logger.Println("Attempting to read data")

	onar.reading = true
	onar.readval = onar.val

	writeMessage := &protocol.ONARWriteMessage{
		Timestamp: onar.timestamp,
		Val:       onar.val,
	}

	// Send write message
	rawMsg, err := proto.Marshal(writeMessage)
	if err != nil {
		return err
	}

	onar.beb.Broadcast(correlationID, rawMsg)
	return nil
}

func (onar *ONAR) Write(val int64) error {

	writeMessage := &protocol.ONARWriteMessage{
		Timestamp: onar.timestamp + 1,
		Val:       val,
	}

	// Send write message
	rawMsg, err := proto.Marshal(writeMessage)
	if err != nil {
		return err
	}

	onar.beb.Broadcast(correlationID, rawMsg)
	return nil
}

func (onar *ONAR) handleAckMessage(conn net.Conn) {

	defer conn.Close()

	onar.logger.Println("Received ack message from host: ", conn.RemoteAddr().(*net.TCPAddr).IP.String())

	ack, err := ioutil.ReadAll(conn)
	if err != nil {
		return
	}

	// Check whether the ackbytes are OK
	if !(bytes.Equal(ack, ackBytes[:])) {
		return
	}

	onar.stateLock.Lock()
	defer onar.stateLock.Unlock()

	// Add the caller to the writeset
	onar.writeset[conn.RemoteAddr().(*net.TCPAddr).IP.String()] = true

	onar.tryDelivering()
}

func (onar *ONAR) handleProcessCrash() {

	for {
		crashedHost := <-onar.pfdOnCrashedListener

		onar.stateLock.Lock()

		// Remove the crashed process from the crashed processes list
		for i := 0; i < len(onar.correctProcesses); i++ {
			if onar.correctProcesses[i] == crashedHost {
				onar.correctProcesses[i] = onar.correctProcesses[len(onar.correctProcesses)-1]
				onar.correctProcesses = onar.correctProcesses[:len(onar.correctProcesses)-1]
				break
			}
		}

		onar.tryDelivering()
		onar.stateLock.Unlock()
	}

}

func (onar *ONAR) handleBebDeliver() {

	for {
		bebMsg := <-onar.bebOnDeliverListener

		// Ignore unintended messages
		if bebMsg.CorrelationID != correlationID {
			continue
		}

		// Get underlying ONARWriteMessage
		onarMsg := &protocol.ONARWriteMessage{}
		err := proto.Unmarshal(bebMsg.Message, onarMsg)
		if err != nil {
			continue
		}

		onar.logger.Println("Received ONAR Write message")

		if onarMsg.GetTimestamp() > onar.timestamp {
			onar.timestamp = onarMsg.GetTimestamp()
			onar.val = onarMsg.GetVal()
		}

		// Send ack message to the ONAR process
		err = util.SendMessage(bebMsg.SourceHost, onar.onarPort, ackBytes[:])
		if err != nil {
			continue
		}
	}

}

func (onar *ONAR) tryDelivering() {

	// Check whether all correct processes are part of the writeset
	// If it is the case, end the current write or read operation
	if onar.allCorrectHaveAcked() {
		onar.writeset = make(map[string]bool, 0)
		if onar.reading {
			onar.reading = false
			onar.onReadReturn.Submit(ReadReturn{value: onar.readval})
		} else {
			onar.onWriteReturn.Submit(WriteReturn{})
		}
	}
}

func (onar *ONAR) allCorrectHaveAcked() bool {

	for _, correctHost := range onar.correctProcesses {
		if !onar.writeset[correctHost] {
			return false
		}
	}

	return true
}
