package lib

import (
	"sync"
)

type Process struct {
	env        *Environment
	resumeChan chan struct{}
	host       *Host

	name         string
	noMoreEvents bool
	data         interface{}

	Done chan struct{}

	pid uint64
}

type ProcessID struct {
	id uint64
	mutex sync.Mutex
}

func (pid *ProcessID) Next() uint64 {
	pid.mutex.Lock()
	defer pid.mutex.Unlock()
	pid.id++
	return pid.id
}

func ProcWrapper(processStrategy func(*Process, []string), w *Process, args []string) {
	go func() {
		<-w.resumeChan
		processStrategy(w, args)
		//end of process
		w.env.mutex.Lock()
		w.noMoreEvents = true
		delete(env.workers, w.pid)
		w.env.mutex.Unlock()
		w.env.stepEnd <- struct{}{}
	}()
}

func (process *Process) Daemonize() {
	env.mutex.Lock()
	defer env.mutex.Unlock()
	env.daemonList = append(env.daemonList, process)
}

func (p *Process) GetData() interface{} {
	return p.data
}

func (p *Process) GetName() interface{} {
	return p.name
}

func (p *Process) GetEnv() *Environment {
	return p.env
}
