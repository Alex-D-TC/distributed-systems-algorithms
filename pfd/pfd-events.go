package pfd

import (
	"fmt"
	"log"
	"os"

	"github.com/alex-d-tc/distributed-systems-algorithms/util"
)

type onProcessCrashedEventManager struct {

	// Event processing internal state
	listeners []chan<- string
	deadHosts []string

	// Event handling
	internalHandler util.EventHandler

	// Logging
	logger *log.Logger
}

func newOnProcessCrashedEventManager() onProcessCrashedEventManager {

	manager := onProcessCrashedEventManager{
		listeners: []chan<- string{},
		deadHosts: []string{},
		logger:    log.New(os.Stdout, "[ProcessCrashedEventHandler]", log.Ldate|log.Ltime),
	}

	manager.internalHandler = util.NewEventHandler(manager.handleEvent)

	return manager
}

func (manager onProcessCrashedEventManager) AddListener(listener chan<- string) {
	manager.internalHandler.Submit(listener)
}

func (manager onProcessCrashedEventManager) Submit(event string) {
	manager.internalHandler.Submit(event)
}

func (manager onProcessCrashedEventManager) handleEvent(ev interface{}) {
	switch ev.(type) {
	case string:
		manager.handleSubmit(ev.(string))
		break
	case chan<- string:
		manager.handleAddListener(ev.(chan<- string))
		break
	default:
		manager.logger.Println(fmt.Sprintf("Invalid event type submitted. Type was %T", ev))
	}
}

func (manager onProcessCrashedEventManager) handleAddListener(listener chan<- string) {
	manager.listeners = append(manager.listeners, listener)
	for _, host := range manager.deadHosts {
		listener <- host
	}
}

func (manager onProcessCrashedEventManager) handleSubmit(event string) {
	manager.deadHosts = append(manager.deadHosts, event)
	for _, listener := range manager.listeners {
		listener <- event
	}
}
