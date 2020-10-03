package main

import (
	"fmt"
	"math/rand"
	"sync"
)

const (
	STOP   = "__P:"
	START  = "__T:"
	DELETE = "__D:"
)

type GRTChannel struct {
	gid     uint64
	name    string
	running bool
	msg     chan string
}

type GRTChannelMap struct {
	mutex    sync.Mutex
	grtchans map[string]*GRTChannel
}

func (grtmap *GRTChannelMap) unregister(name string) error {
	defer grtmap.mutex.Unlock()
	grtmap.mutex.Lock()

	if _, ok := grtmap.grtchans[name]; !ok {
		return fmt.Errorf("goroutine channel not find: %q", name)
	}
	delete(grtmap.grtchans, name)
	return nil
}

func (grtmap *GRTChannelMap) register(name string) error {
	defer grtmap.mutex.Unlock()
	grtmap.mutex.Lock()
	gchan := &GRTChannel{gid: uint64(rand.Int63()),
		name:    name,
		running: false}
	gchan.msg = make(chan string)

	if grtmap.grtchans == nil {
		grtmap.grtchans = make(map[string]*GRTChannel)
	} else if _, ok := grtmap.grtchans[gchan.name]; ok {
		return fmt.Errorf("goroutine channel already defined: %q", gchan.name)
	}
	grtmap.grtchans[gchan.name] = gchan
	return nil
}

func (grtmap *GRTChannelMap) allRunning() {

}
