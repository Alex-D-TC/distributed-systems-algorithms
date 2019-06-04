package uc

import (
	"errors"
	"sync"

	"github.com/golang/protobuf/proto"

	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/beb"
	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/pfd"
	"github.com/alex-d-tc/distributed-systems-algorithms/protocol"
	"github.com/alex-d-tc/distributed-systems-algorithms/util"
	"github.com/alex-d-tc/distributed-systems-algorithms/util/environment"
)

type UniformConsensus struct {

	// Algorithm state
	// Everything apart from the correctProcesses list refreshes when delivering a message
	correctProcesses []string
	round            uint16
	decision         int64
	decided          bool
	proposalSet      map[int64]bool
	receivedFrom     map[string]bool

	proposed     bool
	processCount uint16

	// Algorithm dependencies
	pfd *pfd.PerfectFailureDetector
	beb *beb.BestEffortBroadcast

	// Inbound listeners
	pfdOnProcessCrashedListener <-chan string
	bebOnDeliverListener        <-chan beb.BestEffortBroadcastMessage

	// Outbound events
	onDeliver *onDecidedManager

	// Sync
	stateLock *sync.Mutex
}

func NewUniformConsensus(env *environment.NetworkEnvironment, beb *beb.BestEffortBroadcast, pfd *pfd.PerfectFailureDetector) *UniformConsensus {

	uc := &UniformConsensus{
		processCount:                (uint16)(len(env.GetHosts())),
		correctProcesses:            env.GetHosts(),
		pfd:                         pfd,
		beb:                         beb,
		pfdOnProcessCrashedListener: pfd.AddOnProcessCrashedListener(),
		bebOnDeliverListener:        beb.AddOnDeliverListener(),
		onDeliver:                   newOnDecidedManager(),
		stateLock:                   &sync.Mutex{},
	}

	uc.refreshAlgorithmState()

	go uc.handleBebDeliver()
	go uc.handleProcessCrashed()

	return uc
}

func (uc *UniformConsensus) Propose(val int64) error {

	if uc.proposed {
		return errors.New("The node is currently proposing a value")
	}

	uc.stateLock.Lock()
	defer uc.stateLock.Unlock()

	uc.proposed = true
	uc.proposalSet[val] = true

	proposalMsg := prepareProposalMessage(uc.round, &uc.proposalSet)
	proposalMsgRaw, err := proto.Marshal(&proposalMsg)
	if err != nil {
		return err
	}

	uc.beb.Broadcast(proposalMsgRaw)

	return nil
}

func (uc *UniformConsensus) handleBebDeliver() {

	for {
		bebMsg := <-uc.bebOnDeliverListener

		// Get the underlying proposal message
		proposeMsg := &protocol.UCProposalMessage{}

		err := proto.Unmarshal(bebMsg.Message, proposeMsg)
		if err != nil {
			continue
		}

		// Drop the message if it is not for my round
		if (uint16)(proposeMsg.Round) != uc.round {
			continue
		}

		uc.stateLock.Lock()

		// Mark the beb source as having messaged me, and merge my proposal set with his
		uc.receivedFrom[bebMsg.SourceHost] = true
		for _, proposal := range proposeMsg.ProposalSet {
			uc.proposalSet[proposal] = true
		}

		// Try delivering
		uc.tryDelivering()

		uc.stateLock.Unlock()
	}
}

func (uc *UniformConsensus) handleProcessCrashed() {

	for {
		crashedProcess := <-uc.pfdOnProcessCrashedListener

		uc.stateLock.Lock()

		// Removed the crashed process from the list
		for i := 0; i < len(uc.correctProcesses); i++ {
			if uc.correctProcesses[i] == crashedProcess {
				uc.correctProcesses[i] = uc.correctProcesses[len(uc.correctProcesses)-1]
				uc.correctProcesses = uc.correctProcesses[:len(uc.correctProcesses)-1]
				break
			}
		}

		// Try delivering
		uc.tryDelivering()

		uc.stateLock.Unlock()
	}
}

func (uc *UniformConsensus) tryDelivering() {

	uc.stateLock.Lock()
	defer uc.stateLock.Unlock()

	// If the node has decided, or not all correct processes have reached me, do not try delivering
	if !setContainsKeys(uc.receivedFrom, uc.correctProcesses) || !uc.decided {
		return
	}

	if uc.round == uc.processCount {
		uc.decision = util.MinimumKeyValueOfMap(&uc.proposalSet)
		uc.onDeliver.Submit(UniformConsensusDecision{Value: uc.decision})

		//  Refreshing the full algorithm state is possible at this point
		// since all correct processes have transmitted their proposal sets for this round
		// this effectively resets the algorithm to square one
		uc.refreshAlgorithmState()
	} else {
		uc.round++
		uc.receivedFrom = make(map[string]bool, 0)

		proposalMsg := prepareProposalMessage(uc.round, &uc.proposalSet)
		proposalMsgRaw, err := proto.Marshal(&proposalMsg)
		if err != nil {
			return
		}

		uc.beb.Broadcast(proposalMsgRaw)
	}
}

func setContainsKeys(set map[string]bool, keys []string) bool {

	for _, key := range keys {
		if !set[key] {
			return false
		}
	}

	return true
}

func prepareProposalMessage(round uint16, proposalSet *map[int64]bool) protocol.UCProposalMessage {
	proposalMsg := protocol.UCProposalMessage{}

	proposalSetRaw := make([]int64, len(*proposalSet))
	for k := range *proposalSet {
		proposalSetRaw = append(proposalSetRaw, k)
	}

	proposalMsg.Round = (uint32)(round)
	proposalMsg.ProposalSet = proposalSetRaw

	return proposalMsg
}

func (uc *UniformConsensus) refreshAlgorithmState() {
	uc.round = 1
	uc.decision = 0
	uc.decided = false
	uc.proposalSet = make(map[int64]bool, 0)
	uc.receivedFrom = make(map[string]bool, 0)
	uc.proposed = false
}
