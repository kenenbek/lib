package lib

import (
	"math"
	"fmt"
)

func (worker *Process) SendTask(task *Task, address string) interface{} {
	event := SendEvent{
		Event: &Event{
			timeEnd: math.MaxFloat64,
			worker:  worker,
		},

		size:          task.size,
		remainingSize: task.size,

		sender:  worker.name,
		address: address,

		task: task,
	}

	event.callbacks = append(event.callbacks, event.deleteSelfFromResource)
	env.mutex.Lock()
	worker.env.SendEventsNameMap[address] = append(worker.env.SendEventsNameMap[address], &event)
	env.mutex.Unlock()

	//should wait for an own turn
	worker.env.stepEnd <- struct{}{}
	<-worker.resumeChan
	return nil
}

func (worker *Process) DetachedSendTask(task *Task, address string) interface{} {

	event := SendEvent{
		Event: &Event{
			timeEnd: math.MaxFloat64,
			worker:  worker,
			async:true,
		},

		size:          task.size,
		remainingSize: task.size,

		sender:  worker.name,
		address: address,

		task: task,
	}

	event.callbacks = append(event.callbacks, event.deleteSelfFromResource)

	env.mutex.Lock()
	worker.env.SendEventsNameMap[address] = append(worker.env.SendEventsNameMap[address], &event)
	env.mutex.Unlock()
	return nil
}

func (worker *Process) ReceiveTask(address string) *Task {
	event := ReceiveMock{
		worker:  worker,
		address: address,
	}

	env.mutex.Lock()
	if _, ok := worker.env.ReceiveEventsNameMap[address]; ok  {
		fmt.Println(address)
		panic("multiple listen on the same address")
	}
	worker.env.ReceiveEventsNameMap[address] =  &event
	env.mutex.Unlock()

	worker.env.stepEnd <- struct{}{}
	<-worker.resumeChan

	return event.task
}

func (worker *Process) SIM_wait(waitTime float64) interface{} {

	event := ConstantEvent{
		&Event{
			timeStart: worker.env.currentTime,
			timeEnd:   worker.env.currentTime + waitTime,
			worker:    worker,
		},
	}
	worker.env.mutex.Lock()
	worker.env.queue = append(worker.env.queue, &event)
	worker.env.mutex.Unlock()

	// Wait for an own turn
	worker.env.stepEnd <- struct{}{}
	<-worker.resumeChan
	return nil
}
