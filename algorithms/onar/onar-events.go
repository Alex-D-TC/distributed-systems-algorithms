package onar

import (
	"log"
	"os"

	"github.com/alex-d-tc/distributed-systems-algorithms/util"
)

type ReadReturn struct {
	value int64
}

type onReadReturnManager struct {
	listeners []chan<- ReadReturn

	internalHandler util.EventHandler

	logger *log.Logger
}

func NewOnReadReturnManager() *onReadReturnManager {

	manager := &onReadReturnManager{
		listeners: make([]chan<- ReadReturn, 0),
		logger:    log.New(os.Stdout, "[ONAROnReadReturnManager]", log.Ldate|log.Ltime),
	}

	manager.internalHandler = util.NewEventHandler(manager.handleMessage)

	return manager
}

func (manager *onReadReturnManager) handleMessage(ev interface{}) {
	switch ev.(type) {
	case chan<- ReadReturn:
		manager.listeners = append(manager.listeners, ev.(chan<- ReadReturn))
		break
	case ReadReturn:
		for _, listener := range manager.listeners {
			listener <- ev.(ReadReturn)
		}
	}
}

func (manager *onReadReturnManager) AddListener() <-chan ReadReturn {
	listener := make(chan ReadReturn, 1)
	manager.internalHandler.Submit((chan<- ReadReturn)(listener))
	return listener
}

func (manager *onReadReturnManager) Submit(ev ReadReturn) {
	manager.internalHandler.Submit(ev)
}

type WriteReturn struct {
}

type onWriteReturnManager struct {
	listeners []chan<- WriteReturn

	internalHandler util.EventHandler

	logger *log.Logger
}

func NewOnWriteReturnManager() *onWriteReturnManager {

	manager := &onWriteReturnManager{
		listeners: make([]chan<- WriteReturn, 0),
		logger:    log.New(os.Stdout, "[ONAROnWriteReturnManager]", log.Ldate|log.Ltime),
	}

	manager.internalHandler = util.NewEventHandler(manager.handleMessage)

	return manager
}

func (manager *onWriteReturnManager) handleMessage(ev interface{}) {
	switch ev.(type) {
	case chan<- WriteReturn:
		manager.listeners = append(manager.listeners, ev.(chan<- WriteReturn))
		break
	case WriteReturn:
		for _, listener := range manager.listeners {
			listener <- ev.(WriteReturn)
		}
	}
}

func (manager *onWriteReturnManager) AddListener() <-chan WriteReturn {
	listener := make(chan WriteReturn, 1)
	manager.internalHandler.Submit((chan<- WriteReturn)(listener))
	return listener
}

func (manager *onWriteReturnManager) Submit(ev WriteReturn) {
	manager.internalHandler.Submit(ev)
}
