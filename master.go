package lib

import (
	"fmt"
	"sync"
)

func master(env *Environment, until interface{}, wg *sync.WaitGroup) {
	if until != nil {
		switch until := until.(type) {
		case nil:
			//do nothing
		default:
			untilFloat64 := until.(float64)
			globalStop := ConstantEvent{
				Event: &Event{timeEnd: untilFloat64},
			}
			globalStop.callbacks = append(globalStop.callbacks, env.stopSimulation)
			env.PutEvents(&globalStop)
		}
	}
	// Initial
	var currentEvent EventInterface
	var isWorkerAlive bool
	defer wg.Done()

	env.WaitWorkers()

	env.CreateTransferEvents()
	currentEvent, isWorkerAlive = env.Step()

	for !env.shouldStop {
		if isWorkerAlive {
			env.FindNextWorkers(currentEvent)
			env.SendStartSignalWorkers()
			env.WaitWorkers()
		}
		env.CreateTransferEvents()
		currentEvent, isWorkerAlive = env.Step()
	}
	fmt.Println("end-master")
}

func isWorkerAlive(currentEvent EventInterface) bool {
	alive := false
	workers := currentEvent.getWorkers()
	for i := range workers {
		if workers[i] != nil && !workers[i].noMoreEvents {
			alive = true
		}
	}
	return alive
}
