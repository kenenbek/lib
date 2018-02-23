package main

import (
	"fmt"
	"lib"
)

func sender(p *lib.Process, args []string) {
 	task1 := lib.NewTask("task", 100, 5, nil)
	task2 := lib.NewTask("task", 100, 1, nil)
	task3 := lib.NewTask("task", 100, 10, nil)
	p.DetachedSendTask(task1, "1")
	p.DetachedSendTask(task3, "3")
	p.SIM_wait(2.5)
	p.DetachedSendTask(task2, "2")
}

func receiver(p *lib.Process, args []string) {
		fmt.Println("start listen", lib.SIM_get_clock())
		_ = p.ReceiveTask(args[0])
		fmt.Println("end listen", lib.SIM_get_clock())

}
