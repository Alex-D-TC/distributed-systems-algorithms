package urb

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/beb"
	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/pfd"
	"github.com/alex-d-tc/distributed-systems-algorithms/protocol"
	"github.com/alex-d-tc/distributed-systems-algorithms/util"
	"github.com/alex-d-tc/distributed-systems-algorithms/util/environment"
)

type URB struct {
	hostname         string
	correctProcesses []string
	delivered        map[uint64]bool
	pending          map[string]map[uint64]*protocol.URBMessage
	ack              map[uint64][]string

	onDeliver *onDeliverEventManager

	beb                  *beb.BestEffortBroadcast
	bebOnDeliverListener <-chan beb.BestEffortBroadcastMessage

	pfd                         *pfd.PerfectFailureDetector
	pfdOnProcessCrashedListener <-chan string

	logger *log.Logger

	mutationLock *sync.Mutex
}

func NewURB(env *environment.NetworkEnvironment, beb *beb.BestEffortBroadcast, pfd *pfd.PerfectFailureDetector) *URB {
	urb := &URB{
		hostname:  env.GetHostname(),
		delivered: make(map[uint64]bool),
		pending:   make(map[string]map[uint64]*protocol.URBMessage),
		ack:       make(map[uint64][]string),
		onDeliver: newOnDeliverEventManager(),

		beb: beb,
		pfd: pfd,

		logger:       log.New(os.Stdout, "[URB]", log.Ldate|log.Ltime),
		mutationLock: &sync.Mutex{},
	}

	// Copy the processes list in the correctProcesses list
	urb.correctProcesses = make([]string, 0)
	for _, host := range env.GetHosts() {
		urb.correctProcesses = append(urb.correctProcesses, host)
	}

	// Init the pending map
	for _, src := range env.GetHosts() {
		urb.pending[src] = make(map[uint64]*protocol.URBMessage, 0)
	}

	urb.bebOnDeliverListener = beb.AddOnDeliverListener()
	urb.pfdOnProcessCrashedListener = pfd.AddOnProcessCrashedListener()

	go urb.handleBebDeliver()
	go urb.handlePFDProcessCrashed()

	return urb
}

func (urb *URB) AddOnDeliverListener() <-chan UrbDeliverEvent {
	return urb.onDeliver.AddListener()
}

func (urb *URB) Broadcast(message []byte) error {

	// Prepare the URB Broadcast message
	urbMsg := &protocol.URBMessage{
		Source:  urb.hostname,
		Payload: message,
		Nonce:   uint64(rand.Int63n(time.Now().Unix())),
	}

	urb.mutationLock.Lock()
	defer urb.mutationLock.Unlock()

	// Add the message as pending
	urb.pending[urbMsg.Source][urbMsg.Nonce] = urbMsg

	rawMsg, err := proto.Marshal(urbMsg)
	if err != nil {
		return err
	}

	// BEB Broadcast the message to all parties
	urb.beb.Broadcast(rawMsg)
	return nil
}

func (urb *URB) handleBebDeliver() {

	for {
		bebMsg := <-urb.bebOnDeliverListener

		// Get the underlying URB message
		urbMsg := &protocol.URBMessage{}

		err := proto.Unmarshal(bebMsg.Message, urbMsg)
		if err != nil {
			urb.logger.Printf("Error extracting URB message from BEB message: %s\n", err.Error())
			continue
		}

		// Lock all internal state manipulation
		urb.mutationLock.Lock()

		// If the message is not ack-ed by this host, mark it as ack-ed
		if !util.FindInStringSlice(urb.ack[urbMsg.GetNonce()], bebMsg.SourceHost) {
			urb.ack[urbMsg.GetNonce()] = append(urb.ack[urbMsg.GetNonce()], bebMsg.SourceHost)
		}

		// If the message is not pending, mark it as pending and BEB broadcast it
		if urb.pending[urbMsg.GetSource()][urbMsg.GetNonce()] == nil {
			urb.pending[urbMsg.GetSource()][urbMsg.GetNonce()] = urbMsg
			urb.beb.Broadcast(bebMsg.Message)
		}

		// Try to see if the message is now deliverable
		urb.tryDelivering()

		urb.mutationLock.Unlock()
	}

}

func (urb *URB) handlePFDProcessCrashed() {

	for {
		host := <-urb.pfdOnProcessCrashedListener

		corr := urb.correctProcesses

		urb.logger.Println(fmt.Sprintf("Marking host %s as dead", host))

		// Lock all internal state manipulation
		urb.mutationLock.Lock()

		// Remove process from the correct processes list
		for i := 0; i < len(corr); i++ {
			if corr[i] == host {
				urb.correctProcesses[i] = urb.correctProcesses[len(urb.correctProcesses)-1]
				urb.correctProcesses = urb.correctProcesses[:len(urb.correctProcesses)-1]
				break
			}
		}

		// Try to see if the message is now deliverable
		urb.tryDelivering()

		urb.mutationLock.Unlock()
	}

}

func (urb *URB) tryDelivering() {

	for _, pendingOfHost := range urb.pending {
		for pendingMsg, pendingMsgData := range pendingOfHost {

			// Deliver all messages which can be delivered and weren't already
			if !urb.delivered[pendingMsg] && urb.checkCanDeliver(pendingMsg) {
				urb.onDeliver.Submit(UrbDeliverEvent{
					Source:  pendingMsgData.GetSource(),
					Message: pendingMsgData.GetPayload(),
				})
			}
		}
	}
}

// checkCanDeliver evaluates whether all currently correct processes have ack-ed the particular message
func (urb *URB) checkCanDeliver(msg uint64) bool {

	sort.Strings(urb.ack[msg])
	sort.Strings(urb.correctProcesses)

	// Check whether all correct processes have acknowledged
	iC := 0
	for iA := 0; iA < len(urb.ack[msg]) && iC < len(urb.correctProcesses); iA++ {
		if urb.ack[msg][iA] == urb.correctProcesses[iC] {
			iC++
		}
	}

	return iC == len(urb.correctProcesses)
}
