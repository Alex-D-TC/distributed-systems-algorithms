package uc

import (
	"errors"
	"log"
	"os"
	"sync"

	"github.com/golang/protobuf/proto"

	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/beb"
	"github.com/alex-d-tc/distributed-systems-algorithms/algorithms/pfd"
	"github.com/alex-d-tc/distributed-systems-algorithms/protocol"
	"github.com/alex-d-tc/distributed-systems-algorithms/util"
	"github.com/alex-d-tc/distributed-systems-algorithms/util/environment"
)

type cachedProposal struct {
	source   string
	proposal *protocol.UCProposalMessage
}

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
	earlyBirds   []cachedProposal

	// Algorithm dependencies
	pfd *pfd.PerfectFailureDetector
	beb *beb.BestEffortBroadcast

	// Inbound listeners
	pfdOnProcessCrashedListener <-chan string
	bebOnDeliverListener        <-chan beb.BestEffortBroadcastMessage

	// Outbound events
	onDecided *onDecidedManager

	// Sync
	stateLock *sync.Mutex

	// Logging
	logger *log.Logger
}

func NewUniformConsensus(env *environment.NetworkEnvironment, beb *beb.BestEffortBroadcast, pfd *pfd.PerfectFailureDetector) *UniformConsensus {

	uc := &UniformConsensus{
		proposed:                    false,
		processCount:                (uint16)(len(env.GetHosts())),
		correctProcesses:            env.GetHosts(),
		pfd:                         pfd,
		beb:                         beb,
		pfdOnProcessCrashedListener: pfd.AddOnProcessCrashedListener(),
		bebOnDeliverListener:        beb.AddOnDeliverListener(),
		onDecided:                   newOnDecidedManager(),
		stateLock:                   &sync.Mutex{},
		logger:                      log.New(os.Stdout, "[UC]", log.Ldate|log.Ltime|log.Lmicroseconds),
	}

	uc.refreshAlgorithmState()

	go uc.handleBebDeliver()
	go uc.handleProcessCrashed()

	return uc
}

func (uc *UniformConsensus) AddOnDecidedListener() <-chan UniformConsensusDecision {
	return uc.onDecided.AddListener()
}

func (uc *UniformConsensus) Propose(val int64) error {

	uc.logger.Println("Attempting to propose value")

	if uc.proposed {
		return errors.New("The node is currently proposing a value")
	}

	uc.stateLock.Lock()
	defer uc.stateLock.Unlock()

	uc.proposed = true
	uc.proposalSet[val] = true

	proposalMsg := prepareProposalMessage(uc.round, uc.proposalSet)
	proposalMsgRaw, err := proto.Marshal(&proposalMsg)
	if err != nil {
		return err
	}

	uc.beb.Broadcast(proposalMsgRaw)

	uc.logger.Printf("Proposed value %d. Awaiting other proposals\n", val)

	return nil
}

func (uc *UniformConsensus) handleBebDeliver() {

	for {
		bebMsg := <-uc.bebOnDeliverListener

		// Wait for any state changes to occur before accepting any other beb requests
		uc.stateLock.Lock()

		// Get the underlying proposal message
		proposeMsg := &protocol.UCProposalMessage{}

		err := proto.Unmarshal(bebMsg.Message, proposeMsg)
		if err != nil {
			uc.stateLock.Unlock()
			continue
		}

		if len(proposeMsg.ProposalSet) == 0 {
			// Ignore malformed messages
			uc.stateLock.Unlock()
			continue
		}

		early := uc.processMessage(bebMsg.SourceHost, proposeMsg)
		if early {
			// Store the early bird for later use
			uc.earlyBirds = append(uc.earlyBirds, cachedProposal{source: bebMsg.SourceHost, proposal: proposeMsg})
		}

		uc.stateLock.Unlock()
	}
}

func (uc *UniformConsensus) tryProcessingEarlyBirds() {

	unprocessed := make([]cachedProposal, 0)

	for _, toProcess := range uc.earlyBirds {
		early := uc.processMessage(toProcess.source, toProcess.proposal)
		if early {
			unprocessed = append(unprocessed, toProcess)
		}
	}

	uc.earlyBirds = unprocessed
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

	// If the node has decided, or not all correct processes have reached me, do not try delivering
	if !setContainsKeys(uc.receivedFrom, uc.correctProcesses) {
		//uc.logger.Println("Not all correct processes have contacted me yet")
		return
	}

	if uc.decided {
		//uc.logger.Println("The node has already decided. No need to try deciding again")
		return
	}

	if uc.round == uc.processCount {
		uc.decision = util.MinimumKeyValueOfMap(&uc.proposalSet)
		uc.onDecided.Submit(UniformConsensusDecision{Value: uc.decision})

		uc.logger.Printf("Successfully completed the final round. Deciding on %d\n", uc.decision)

		//  Refreshing the full algorithm state is possible at this point
		// since all correct processes have transmitted their proposal sets for this round
		// this effectively resets the algorithm to square one
		uc.refreshAlgorithmState()
	} else {
		uc.round++
		uc.receivedFrom = make(map[string]bool)

		uc.logger.Printf("Successfully completed round %d. Advancing\n", uc.round-1)

		// Try processing early birds before sending my message
		uc.tryProcessingEarlyBirds()

		proposalMsg := prepareProposalMessage(uc.round, uc.proposalSet)
		proposalMsgRaw, err := proto.Marshal(&proposalMsg)
		if err != nil {
			return
		}

		uc.beb.Broadcast(proposalMsgRaw)
	}
}

func (uc *UniformConsensus) processMessage(source string, proposal *protocol.UCProposalMessage) (early bool) {
	uc.logger.Printf("Received proposal from %s\n", source)
	early = false

	// Drop the message if it is not for my round
	if (uint16)(proposal.Round) != uc.round {
		uc.logger.Printf("Proposal was for round %d, when at round %d\n", proposal.Round, uc.round)
		early = true
		return
	}

	uc.logger.Printf("Received proposal set %v from %s\n", proposal.ProposalSet, source)

	// Mark the beb source as having messaged me, and merge my proposal set with his
	uc.receivedFrom[source] = true
	for _, proposal := range proposal.ProposalSet {
		uc.proposalSet[proposal] = true
	}

	// Try delivering
	uc.tryDelivering()
	return
}

func setContainsKeys(set map[string]bool, keys []string) bool {

	for _, key := range keys {
		if !set[key] {
			return false
		}
	}

	return true
}

func prepareProposalMessage(round uint16, proposalSet map[int64]bool) protocol.UCProposalMessage {
	proposalMsg := protocol.UCProposalMessage{}

	proposalSetRaw := make([]int64, 0, len(proposalSet))
	for k := range proposalSet {
		if proposalSet[k] {
			proposalSetRaw = append(proposalSetRaw, k)
		}
	}

	proposalMsg.Round = (uint32)(round)
	proposalMsg.ProposalSet = proposalSetRaw

	return proposalMsg
}

func (uc *UniformConsensus) refreshAlgorithmState() {
	uc.round = 1
	uc.decision = 0
	uc.decided = false
	uc.proposalSet = make(map[int64]bool)
	uc.receivedFrom = make(map[string]bool)
	uc.proposed = false
	uc.earlyBirds = make([]cachedProposal, 0)
}
