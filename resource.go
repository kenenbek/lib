package lib

import (
	"sort"
	"sync"
	"sync/atomic"
)

type Resource struct {
	queue     []*TransferEvent
	bandwidth float64

	lastTimeRequest float64
	mutex           sync.Mutex

	counter  int64
	_counter int64
	env      *Environment
}

func (r *Resource) Put(e *TransferEvent) {
	atomic.AddInt64(&r.counter, 1)
	atomic.AddInt64(&r._counter, -1)
	r.queue = append(r.queue, e)
}

func (r *Resource) EstimateTimeEnd(e *SendEvent) {
	qLen := len(r.queue)
	if qLen == 0 {
		e.timeEnd = r.env.currentTime + e.remainingSize/(r.bandwidth/float64(r._counter))
	} else if e.remainingSize < r.queue[0].sendEvent.remainingSize {
		e.timeEnd = r.env.currentTime + e.remainingSize/(r.bandwidth/float64(r.counter+r._counter))
	}
}

func (env *Environment) FindNextTransferEvent() {
	for route := range env.routesMap {
		sort.Sort(ByRemainingSize(env.routesMap[route].queue))
		if len(env.routesMap[route].queue) > 0 {
			env.routesMap[route].queue[0].sendEvent.timeEnd = env.currentTime + env.routesMap[route].queue[0].sendEvent.remainingSize/(env.routesMap[route].bandwidth/float64(env.routesMap[route].counter))
		}
		env.routesMap[route]._counter = 0
	}
}

type RoutesMap map[Route]*Link

func (routeMap RoutesMap) Get(r Route) (d *Link) {
	d, ok := routeMap[r]
	if ok {
		return
	}
	d, ok = routeMap[Route{start: r.finish, finish: r.start}]
	if ok {
		return
	}
	return nil
}

type Route struct {
	start  *Host
	finish *Host
}

type Link struct {
	*Resource
	name string
}