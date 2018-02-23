package lib

import (
	"sync/atomic"
)

type Host struct {
	name      string
	processes []*Process
	speed     float64
	storage *Storage
}

func (env *Environment) getHostByName(name string) *Host {
	return env.hostsMap[name]
}

func (process *Process) GetHost() *Host {
	return process.host
}

func (host *Host) GetName() string {
	return host.name
}

func (host *Host) GetProcessesAmount() int {
	return len(host.processes)
}

func (host *Host) GetStorage() *Storage {
	return host.storage
}


type StorageType struct {
	typeId string
	writeRate float64
	readRate float64
	size float64
}

type Storage struct {
	*StorageType
	name string
	usedSize int64
}


func (storage *Storage) WritePacketSize(){
	atomic.AddInt64(&(storage.usedSize), 1)
}

func (storage *Storage) DeletePacketSize(){
	atomic.AddInt64(&(storage.usedSize), -1)
}