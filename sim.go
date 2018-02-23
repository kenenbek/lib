package lib

import (
	"strings"
	"sync"
	"sync/atomic"
)

func SIM_init() {
	env = NewEnvironment()
}

func SIM_run(until interface{}) {
	var wg sync.WaitGroup
	wg.Add(1)
	go master(env, until, &wg)
	wg.Wait()
}

func SIM_get_clock() float64 {
	return env.currentTime
}

func SIM_process_create(name string, f func(*Process, []string), host *Host, data interface{}) *Process {
	args := make([]string, 0, 0)
	return SIM_process_create_with_agrs(name, f, host, data, args)
}

func SIM_process_create_with_agrs(name string, f func(*Process, []string), host *Host, data interface{}, args []string) *Process {
	pid := env.pid.Next()
	p := &Process{
		name:         name,
		env:          env,
		resumeChan:   make(chan struct{}),
		host:         host,
		noMoreEvents: false,
		Done:         make(chan struct{}),
		data:         data,
		pid:pid,
	}
	ProcWrapper(f, p, args)
	env.mutex.Lock()
	env.workers[pid] = p
	env.mutex.Unlock()
	atomic.AddUint64(&env.waitWorkerAmount, 1)
	return p
}

func SIM_function_register(FuncName string, Func func(*Process, []string)) {
	env.FunctionsMap[FuncName] = Func
}

func SIM_get_servers() []*Host {
	var servers []*Host
	for key := range env.hostsMap {
		if strings.HasPrefix(key, "Server") {
			servers = append(servers, env.hostsMap[key])
		}
	}
	return servers
}
