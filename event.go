package lib

import (
	"fmt"
	"math"
)

type EventInterface interface {
	update(float64)
	calculateTimeEnd()
	getWorkers() []*Process
	getTimeEnd() float64
	getSize() float64
	getCallbacks() []func(EventInterface)
}

type Event struct {
	id        string
	timeStart float64
	timeEnd   float64

	callbacks []func(EventInterface)
	worker    *Process
	async  bool
}

type ConstantEvent struct {
	*Event
}

type TransferEvent struct {
	sendEvent    *SendEvent
	receiveEvent *ReceiveMock
}

type SendEvent struct {
	*Event
	size float64

	remainingSize float64
	resource      interface{}
	sender        string
	address       string

	task *Task
}

type ReceiveMock struct {
	worker  *Process
	address string

	task *Task
}

func (e *ConstantEvent) update(_ float64) {
	//Nothing to update because event is immutable
}

func (e *ConstantEvent) calculateTimeEnd() {

}

func (e *ConstantEvent) getWorkers() []*Process {
	return []*Process{e.worker}
}

func (e *ConstantEvent) getTimeEnd() float64 {
	return e.timeEnd
}

func (e *ConstantEvent) getSize() float64 {
	return 0
}

func (e *ConstantEvent) getCallbacks() []func(EventInterface) {
	return e.callbacks
}

func (e *TransferEvent) update(deltaTime float64) {
	event := e.sendEvent
	event.remainingSize -= deltaTime * (event.resource.(*Link).bandwidth / float64(event.resource.(*Link).counter))
}

func (e *TransferEvent) calculateTimeEnd() {
	event := e.sendEvent
	event.timeEnd = event.worker.env.currentTime + event.remainingSize/(event.resource.(*Link).bandwidth/float64(event.resource.(*Link).counter))
}

func (e *TransferEvent) getWorkers() []*Process {
	return []*Process{e.sendEvent.worker, e.receiveEvent.worker}
}

func (e *TransferEvent) getTimeEnd() float64 {
	return e.sendEvent.timeEnd
}

func (e *TransferEvent) getSize() float64 {
	return e.sendEvent.remainingSize
}

func (e *TransferEvent) getCallbacks() []func(EventInterface) {
	return e.sendEvent.callbacks
}

type ByTime []EventInterface

func (s ByTime) Len() int {
	return len(s)
}
func (s ByTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByTime) Less(i, j int) bool {
	if s[i].getTimeEnd() == math.MaxFloat64 && s[j].getTimeEnd() == math.MaxFloat64 {
		return s[i].getSize() < s[j].getSize()
	} else if s[i].getTimeEnd() == math.MaxFloat64 {
		return false
	} else if s[j].getTimeEnd() == math.MaxFloat64 {
		return true
	} else {
		return s[i].getTimeEnd() < s[j].getTimeEnd()
	}
}

type ByRemainingSize []*TransferEvent

func (s ByRemainingSize) Len() int {
	return len(s)
}

func (s ByRemainingSize) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByRemainingSize) Less(i, j int) bool {
	return s[i].sendEvent.remainingSize < s[j].sendEvent.remainingSize
}

type ByTransferTime []*SendEvent

func (s ByTransferTime) Len() int {
	return len(s)
}
func (s ByTransferTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByTransferTime) Less(i, j int) bool {
	return s[i].timeEnd < s[j].timeEnd
}

func (e *Event) String() string {
	return fmt.Sprintf("Start %v | End  %v\n", e.timeStart, e.timeEnd)
}


func(_ *SendEvent) deleteSelfFromResource(eventInterface EventInterface){
	if CE, ok := eventInterface.(*TransferEvent); ok {
		CE.sendEvent.resource.(*Link).CounterMinus()
		_, CE.sendEvent.resource.(*Link).queue = CE.sendEvent.resource.(*Link).queue[0], CE.sendEvent.resource.(*Link).queue[1:]}
}