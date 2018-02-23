package lib

import (
	"fmt"
)

func master(env *Environment, until interface{}) {
	if until != nil {
		untilFloat64 := until.(float64)
		globalStop := ConstantEvent{
			Event: &Event{timeEnd: untilFloat64},
		}
		globalStop.callbacks = append(globalStop.callbacks, env.stopSimulation)
		env.queue = append(env.queue, &globalStop)
	}

	var currentEvent EventInterface

	for !env.shouldStop {
		env.FindNextWorkers(currentEvent)
		env.SendStartToSignalWorkers()
		env.WaitWorkers()
		env.CreateTransferEvents()
		currentEvent = env.Step()
	}
	fmt.Println("end-master")
}
