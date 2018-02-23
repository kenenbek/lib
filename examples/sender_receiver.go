package main

import (
	"lib"
	"fmt"
)

func sender(p *lib.Process, args []string){
	task := lib.NewTask("task", 100, 100, nil)
	p.SendTask(task, "receiver")
}


func receiver(p *lib.Process, args []string){
	fmt.Println("start listen", lib.SIM_get_clock())
	_ = p.ReceiveTask("receiver")
	fmt.Println("end listen", lib.SIM_get_clock())
}
