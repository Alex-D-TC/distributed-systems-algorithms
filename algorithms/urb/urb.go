package urb

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/golang/protobuf/proto"

	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/beb"
	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/pfd"
	"github.com/alex-d-tc/distributed-systems-algorithms/protocol"
	"github.com/alex-d-tc/distributed-systems-algorithms/util/environment"
)

type URB struct {
	hostname         string
	correctProcesses []string
	delivered        map[int64]bool
	pending          map[string]map[int64]protocol.URBMessage
	ack              map[int64][]string

	onDeliver *onDeliverEventManager

	beb                  *beb.BestEffortBroadcast
	bebOnDeliverListener <-chan beb.BestEffortBroadcastMessage

	pfd                         *pfd.PerfectFailureDetector
	pfdOnProcessCrashedListener <-chan string

	logger *log.Logger
}

func NewURB(env *environment.NetworkEnvironment, beb *beb.BestEffortBroadcast, pfd *pfd.PerfectFailureDetector) *URB {
	urb := &URB{
		hostname:  env.GetHostname(),
		delivered: make(map[int64]bool),
		pending:   make(map[string]map[int64]protocol.URBMessage),
		ack:       make(map[int64][]string),
		onDeliver: newOnDeliverEventManager(),

		beb: beb,
		pfd: pfd,

		logger: log.New(os.Stdout, "[URB]", log.Ldate|log.Ltime),
	}

	// Copy the processes list in the correctProcesses list
	urb.correctProcesses = make([]string, 0)
	for _, host := range env.GetHosts() {
		urb.correctProcesses = append(urb.correctProcesses, host)
	}

	urb.bebOnDeliverListener = beb.AddOnDeliverListener()
	urb.pfdOnProcessCrashedListener = pfd.AddOnProcessCrashedListener()

	go urb.handleBebDeliver()
	go urb.handlePFDProcessCrashed()

	return urb
}

// checkCanDeliver evaluates whether all correct messages have ack-ed the particular message
func (urb *URB) checkCanDeliver(msg int64) bool {

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

func (urb *URB) tryDelivering() {

	delivered := make([](struct {
		string
		int64
	}), 0)

	for host, pendingOfHost := range urb.pending {
		for pendingMsg, pendingMsgData := range pendingOfHost {

			if urb.checkCanDeliver(pendingMsg) {
				// Deliver all messages which can be delivered
				urb.onDeliver.Submit(UrbDeliverEvent{
					Source:  pendingMsgData.GetSource(),
					Message: pendingMsgData.GetPayload(),
				})

				delivered = append(delivered, struct {
					string
					int64
				}{host, pendingMsg})
			}
		}
	}

	// Remove all delivered messages
	for _, deliv := range delivered {
		delete(urb.pending[deliv.string], deliv.int64)
	}
}

func (urb *URB) Broadcast(message []byte) error {
	// Prepare the URB Broadcast message
	urbMsg := &protocol.URBMessage{
		Source:  urb.hostname,
		Payload: message,
	}

	rawMsg, err := proto.Marshal(urbMsg)
	if err != nil {
		return err
	}

	// BEB Broadcast the message to all parties
	urb.Broadcast(rawMsg)
	return nil
}

func (urb *URB) handleBebDeliver() {

}

func (urb *URB) handlePFDProcessCrashed() {

	for {
		host := <-urb.pfdOnProcessCrashedListener

		corr := urb.correctProcesses

		urb.logger.Println(fmt.Sprintf("Marking host %s as dead", host))

		// Remove process from the correct processes list
		for i := 0; i < len(corr); i++ {
			if corr[i] == host {
				urb.correctProcesses[i] = urb.correctProcesses[len(urb.correctProcesses)-1]
				urb.correctProcesses = urb.correctProcesses[:len(urb.correctProcesses)-1]
				break
			}
		}
	}

}
