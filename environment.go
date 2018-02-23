package lib

import (
	"sort"
	"sync"
	"sync/atomic"
	_"fmt"
)

var env *Environment


type Environment struct {
	currentTime float64
	workers     map[uint64]*Process
	routesMap   RoutesMap
	queue       []EventInterface
	mutex       sync.Mutex
	shouldStop  bool
	hostsMap    map[string]*Host
	storagesMap map[string]*Storage

	SendEventsNameMap map[string][]*SendEvent
	ReceiveEventsNameMap map[string]*ReceiveMock
	ReceiverSendersMap map[*ReceiveMock][]*SendEvent

	FunctionsMap map[string]func(*Process, []string)
	daemonList   []*Process

	pid ProcessID

	waitWorkerAmount uint64
	stepEnd chan interface{}
	nextWorkers []*Process
}

func NewEnvironment() *Environment{
	e := &Environment{
		workers:make(map[uint64]*Process),
		SendEventsNameMap:make(map[string][]*SendEvent),
		ReceiveEventsNameMap:make(map[string]*ReceiveMock),
		ReceiverSendersMap:make(map[*ReceiveMock][]*SendEvent),

		stepEnd: make(chan interface{}),
	}
	return e
}

func (env *Environment) stopSimulation(_ EventInterface) {
	env.shouldStop = true
}

func (env *Environment) updateQueue(deltaTime float64) {
	// Some amount of data has been sent over time

	for index := range env.queue {
		event := env.queue[index]
		event.update(deltaTime)
	}

}

func (env *Environment) CreateTransferEvents() {
	for key := range env.SendEventsNameMap{
		if _, ok := env.ReceiveEventsNameMap[key]; ok{
			env.ReceiverSendersMap[env.ReceiveEventsNameMap[key]] = env.SendEventsNameMap[key]
		}
	}

	// Оценить количество сетей, которые будут использоваться при передаче данных
	for receiveMock := range env.ReceiverSendersMap {
		SendQueue := env.ReceiverSendersMap[receiveMock]
		for i := range SendQueue {
			sendEvent := SendQueue[i]
			route := Route{start: receiveMock.worker.host, finish: sendEvent.worker.host}
			resource := env.routesMap.Get(route)
			sendEvent.resource = resource
			resource._counter++
		}
	}

	for receiveMock := range env.ReceiverSendersMap {
		SendQueue := env.ReceiverSendersMap[receiveMock]
		for i := range SendQueue {
			sendEvent := SendQueue[i]
			sendEvent.resource.(*Link).EstimateTimeEnd(sendEvent)
		}
		sort.Sort(ByTransferTime(env.ReceiverSendersMap[receiveMock]))

		// If there exists pair of receive event and send one then create TransferEvent
		if len(env.ReceiverSendersMap[receiveMock]) > 0 {
			sendEvent := env.ReceiverSendersMap[receiveMock][0]
			sendEvent.timeStart = env.currentTime

			TRANSFEREVENT := TransferEvent{
				receiveEvent: receiveMock,
				sendEvent:    sendEvent,
			}

			receiveMock.task = sendEvent.task

			sendEvent.resource.(*Link).Put(&TRANSFEREVENT)
			env.queue = append(env.queue, &TRANSFEREVENT)

			// DELETE ReceiverSendersMap
			delete(env.SendEventsNameMap, sendEvent.address)
			delete(env.ReceiveEventsNameMap, receiveMock.address)
			delete(env.ReceiverSendersMap, receiveMock)
		}
	}
	env.FindNextTransferEvent()
}

func (env *Environment) Step() EventInterface {
	// check daemons
	if len(env.workers) == 0 || len(env.workers) == len(env.daemonList) {
		env.shouldStop = true
		return nil
	}
	if len(env.queue) == 0 {
		panic("deadlock")
	}
	var currentEvent EventInterface
	sort.Sort(ByTime(env.queue))

	currentEvent, env.queue = env.queue[0], env.queue[1:]

	env.updateQueue(currentEvent.getTimeEnd() - env.currentTime)
	env.currentTime = currentEvent.getTimeEnd()

	// Process the event callbacks
	callbacks := currentEvent.getCallbacks()
	for _, callback := range callbacks {
		callback(currentEvent)
	}

	return currentEvent
}


func (env *Environment) FindNextWorkers(event EventInterface){
	env.nextWorkers = nil
	switch event.(type) {
	case *TransferEvent:
		te := event.(*TransferEvent)

		if !te.sendEvent.worker.noMoreEvents && !te.sendEvent.async {
			env.nextWorkers = append(env.nextWorkers, te.sendEvent.worker)
		}
		if !te.receiveEvent.worker.noMoreEvents {
			env.nextWorkers = append(env.nextWorkers, te.receiveEvent.worker)
		}

	case *ConstantEvent:
		ce := event.(*ConstantEvent)
		if !ce.worker.noMoreEvents {
			env.nextWorkers = append(env.nextWorkers, ce.worker)
		}
	case nil:
		for key := range env.workers{
			env.nextWorkers = append(env.nextWorkers, env.workers[key])
		}
	}

}

func (env *Environment) SendStartToSignalWorkers() {
	atomic.StoreUint64(&env.waitWorkerAmount, uint64(len(env.nextWorkers)))
	for key := range env.nextWorkers {
		env.nextWorkers[key].resumeChan <- struct{}{}
	}
}

func (env *Environment) WaitWorkers() {
	remaining := uint64(0)
	for atomic.LoadUint64(&env.waitWorkerAmount) != remaining {
		<- env.stepEnd
		remaining++
	}

	atomic.StoreUint64(&env.waitWorkerAmount, 0)
}
