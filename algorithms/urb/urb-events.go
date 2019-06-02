package urb;

import (
	"log"
	"os"

	"github.com/alex-d-tc/distributed-systems-algorithms/util"
)

type UrbDeliverEvent struct {
	Source string
	Message []byte
}

type onDeliverEventManager struct {

	// Listener internal state
	listeners []chan<-UrbDeliverEvent

	// Internal event handler
	internalHandler util.EventHandler

	// Logging
	logger *log.Logger
}

func newOnDeliverEventManager() *onDeliverEventManager {
	manager := &onDeliverEventManager{
		listeners: make([]chan<-UrbDeliverEvent, 0),
		logger: log.New(os.Stdout, "[URBOnDeliverManager]", log.Ldate|log.Ltime),
	}

	manager.internalHandler = util.NewEventHandler(manager.handleSubmit)

	return manager
}

func (manager *onDeliverEventManager) AddListener() <-chan UrbDeliverEvent {

	listener := make(chan UrbDeliverEvent, 1)
	manager.internalHandler.Submit(listener)

	return listener
}

func (manager *onDeliverEventManager) Submit(event UrbDeliverEvent) {
	manager.internalHandler.Submit(event)
}

func (manager *onDeliverEventManager) handleSubmit(ev interface{}) {

	switch ev.(type) {
	case chan UrbDeliverEvent:
		manager.logger.Println("Added a listener")
		manager.listeners = append(manager.listeners, ev.(chan<-UrbDeliverEvent))
		break
	case UrbDeliverEvent:
		manager.logger.Println("Delivering URB message")
		for _, listener := range manager.listeners {
			listener<-ev.(UrbDeliverEvent)
		}
		break
	}
}
