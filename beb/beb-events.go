package beb

import (
	"fmt"
	"log"
	"os"

	"github.com/alex-d-tc/distributed-systems-algorithms/util"
)

type onDeliverManager struct {
	listeners []chan<- BestEffortBroadcastMessage

	// Internal event handler
	internalHandler util.EventHandler

	// Logging
	logger *log.Logger
}

func newOnDeliverManager() onDeliverManager {

	manager := onDeliverManager{
		logger: log.New(os.Stdout, "[BebOnDeliverManager]", log.Ldate|log.Ltime),
	}

	manager.internalHandler = util.NewEventHandler(manager.handleEvent)

	return manager
}

func (manager onDeliverManager) AddListener() <-chan BestEffortBroadcastMessage {

	listener := make(chan BestEffortBroadcastMessage, 1)
	manager.internalHandler.Submit(listener)

	return listener
}

func (manager onDeliverManager) Submit(message BestEffortBroadcastMessage) {
	manager.internalHandler.Submit(message)
}

func (manager onDeliverManager) handleEvent(ev interface{}) {
	switch ev.(type) {
	case chan BestEffortBroadcastMessage:
		manager.handleAddListener(ev.(chan BestEffortBroadcastMessage))
		break
	case BestEffortBroadcastMessage:
		manager.handleOnDeliver(ev.(BestEffortBroadcastMessage))
		break
	default:
		manager.logger.Println(fmt.Sprintf("Invalid event type submitted. Type was %T", ev))
	}
}

func (manager onDeliverManager) handleAddListener(listener chan<- BestEffortBroadcastMessage) {
	manager.listeners = append(manager.listeners, listener)
}

func (manager onDeliverManager) handleOnDeliver(message BestEffortBroadcastMessage) {
	for _, listener := range manager.listeners {
		listener <- message
	}
}
