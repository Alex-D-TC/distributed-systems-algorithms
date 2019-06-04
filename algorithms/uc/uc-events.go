package uc

import (
	"fmt"
	"log"
	"os"

	"github.com/alex-d-tc/distributed-systems-algorithms/util"
)

type UniformConsensusDecision struct {
	Value int64
}

type onDecidedManager struct {
	listeners []chan<- UniformConsensusDecision

	// Internal event handler
	internalHandler util.EventHandler

	// Logging
	logger *log.Logger
}

func newOnDecidedManager() *onDecidedManager {

	manager := &onDecidedManager{
		logger: log.New(os.Stdout, "[UCOnDeliverManager]", log.Ldate|log.Ltime),
	}
	manager.internalHandler = util.NewEventHandler(manager.handleEvent)

	return manager
}

func (manager *onDecidedManager) AddListener() <-chan UniformConsensusDecision {

	listener := make(chan UniformConsensusDecision, 1)
	manager.internalHandler.Submit(listener)

	return listener
}

func (manager *onDecidedManager) Submit(message UniformConsensusDecision) {
	manager.internalHandler.Submit(message)
}

func (manager *onDecidedManager) handleEvent(ev interface{}) {
	switch ev.(type) {
	case chan UniformConsensusDecision:
		manager.logger.Println("Added listener")
		manager.listeners = append(manager.listeners, ev.(chan<- UniformConsensusDecision))
		break
	case UniformConsensusDecision:
		for _, listener := range manager.listeners {
			listener <- ev.(UniformConsensusDecision)
		}
		break
	default:
		manager.logger.Println(fmt.Sprintf("Invalid event type submitted. Type was %T", ev))
	}
}
