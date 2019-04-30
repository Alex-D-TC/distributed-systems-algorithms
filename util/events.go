package util

type EventHandler struct {
	eventQueue chan interface{}
}

func NewEventHandler(eventHandler func(interface{})) EventHandler {

	handler := EventHandler{
		eventQueue: make(chan interface{}, 1),
	}

	go handler.processEventQueue(eventHandler)

	return handler
}

func (eventHandler EventHandler) processEventQueue(eventProcessor func(interface{})) {
	for {
		ev := <-eventHandler.eventQueue
		eventProcessor(ev)
	}
}

func (eventHandler EventHandler) Submit(ev interface{}) {
	eventHandler.eventQueue <- ev
}
