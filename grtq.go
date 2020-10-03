package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

type GRManager struct {
	GRTChannelMap *GRTChannelMap
	delall        chan string
	wg            *sync.WaitGroup
	pool          *ants.Pool
}

func NewDefaultGRManager() *GRManager {
	gm := &GRTChannelMap{}
	delall := make(chan string)
	var wg = &sync.WaitGroup{}
	pool, _ := ants.NewPool(math.MaxInt32,) //ants.WithExpiryDuration(time.Second * 2))
	return &GRManager{GRTChannelMap: gm, delall: delall, wg: wg, pool: pool}
}

func NewGRManager(numworkers int) *GRManager {
	gm := &GRTChannelMap{}
	delall := make(chan string)
	var wg = &sync.WaitGroup{}
	pool, _ := ants.NewPool(numworkers, ants.WithPreAlloc(true))
	return &GRManager{GRTChannelMap: gm, delall: delall, wg: wg, pool: pool}
}

func (gm *GRManager) StopGRT(name string) error {
	stopChan, ok := gm.GRTChannelMap.grtchans[name]
	if !ok {
		return fmt.Errorf("not found goroutine name :" + name)
	}
	stopChan.msg <- STOP + strconv.Itoa(int(stopChan.gid))
	return nil
}

func (gm *GRManager) StopAllGRT() error {

	for _, stopchan := range gm.GRTChannelMap.grtchans {
		stopchan.msg <- STOP + strconv.Itoa(int(stopchan.gid))
		// stopchan.msg <- STOP + strconv.Itoa(int(stopchan.gid))
	}

	// fmt.Println("Releasing...............................................................")

	// gm.pool.Release()
	// gm.pool.Reboot()
	// gm.wg.Wait()
	return nil
}

func (gm *GRManager) DeleteGRT(name string) error {
	delChan, ok := gm.GRTChannelMap.grtchans[name]
	if !ok {
		return fmt.Errorf("not found goroutine name :" + name)
	}
	delChan.msg <- DELETE + strconv.Itoa(int(delChan.gid))
	return nil
}

func (gm *GRManager) DeleteAllGRT() error {
	go func() {
		for start := time.Now(); (time.Since(start) < (time.Second * 2)) || gm.pool.Running() != 0; {
			fmt.Printf("Termination since %q ago, %d running goroutines still running \n", time.Since(start).String(), gm.pool.Running())
			time.Sleep(time.Millisecond * 200)
		}
	}()
	for gm.pool.Running() != 0 {
		gm.StopAllGRT()
	}
	// gm.pool.Release()
	// gm.pool.Reboot()

	for _, delchan := range gm.GRTChannelMap.grtchans {
		delchan.msg <- DELETE + strconv.Itoa(int(delchan.gid))
	}

	
	// gm.delall <- DELETE
	gm.wg.Wait()

	// fmt.Println(<-gm.delall)
	

	return nil
}

func (gm *GRManager) StartGRT(name string) error {
	startChan, ok := gm.GRTChannelMap.grtchans[name]
	if !ok {
		return fmt.Errorf("not found goroutine name :" + name)
	}
	startChan.msg <- START + strconv.Itoa(int(startChan.gid))
	return nil
}

func (gm *GRManager) StartAllGRT() error {
	for _, strtchan := range gm.GRTChannelMap.grtchans {
		strtchan.msg <- START + strconv.Itoa(int(strtchan.gid))
	}
	return nil
}

func (gm *GRManager) newTask(name string, start bool, fc interface{}, args ...interface{}) {
	go func(gm *GRManager, name string, start bool, fc interface{}, args ...interface{}) {
		// gm.GRTChannelMap.mutex.Lock()
		err := gm.GRTChannelMap.register(name)
		if err != nil {
			return
		}
		for !start {
			select {
			case info := <-gm.delall:
				taskInfo := strings.Split(info, ":")
				signal := taskInfo[0]
				if signal == "__D" {
					fmt.Println("Task [" + gm.GRTChannelMap.grtchans[name].name + "] quit")
					// gm.GRTChannelMap.mutex.Lock()
					gm.GRTChannelMap.unregister(name)
					gm.delall <- DELETE
					return
				} else {
					fmt.Println("Unknown Signal")
				}

			case info := <-gm.GRTChannelMap.grtchans[name].msg:
				taskInfo := strings.Split(info, ":")
				signal, gid := taskInfo[0], taskInfo[1]
				if gid == strconv.Itoa(int(gm.GRTChannelMap.grtchans[name].gid)) {
					if signal == "__P" && gm.GRTChannelMap.grtchans[name].running{
						fmt.Println("Task [" + gm.GRTChannelMap.grtchans[name].name + "] stopped")
						start = false
					} else if signal == "__T" {
						fmt.Println("Task [" + gm.GRTChannelMap.grtchans[name].name + "] started")
						start = true
					} else if signal == "__D" {
						fmt.Println("Task [" + gm.GRTChannelMap.grtchans[name].name + "] quit")
						// gm.GRTChannelMap.mutex.Lock()
						gm.GRTChannelMap.unregister(name)
						return
					} 
					// else {
					// 	fmt.Println("Unknown Signal")
					// }
				}
			default:
				// fmt.Println("No Signal")
				time.Sleep(time.Millisecond * 200)
			}
		}
		// time.Sleep(time.Millisecond*500)
		runn, oky := gm.GRTChannelMap.grtchans[name]
		if oky {
			if start && !runn.running {
				if len(args) > 1 {
					_ = gm.pool.Submit(func() {
						if grtchan, ok := gm.GRTChannelMap.grtchans[name]; ok { // gm.GRTChannelMap.mutex.Lock()
							// time.Sleep(time.Millisecond * 200)
							gm.GRTChannelMap.mutex.Lock()
							grtchan.running = true
							gm.GRTChannelMap.mutex.Unlock()
							gm.wg.Add(1)
							fc.(func(...interface{}))(args)
							gm.wg.Done()
							gm.GRTChannelMap.mutex.Lock()
							grtchan.running = false
							gm.GRTChannelMap.mutex.Unlock()
						} else {
							return
						}
					})
				} else if len(args) == 1 {
					_ = gm.pool.Submit(func() {
						if grtchan, ok := gm.GRTChannelMap.grtchans[name]; ok {
							// time.Sleep(time.Millisecond * 200)
							gm.GRTChannelMap.mutex.Lock()
							grtchan.running = true
							gm.GRTChannelMap.mutex.Unlock()
							gm.wg.Add(1)
							fc.(func(interface{}))(args[0])
							gm.wg.Done()
							gm.GRTChannelMap.mutex.Lock()
							grtchan.running = false
							gm.GRTChannelMap.mutex.Unlock()
						} else {
							return
						}
					})
				} else {
					_ = gm.pool.Submit(func() {
						if grtchan, ok := gm.GRTChannelMap.grtchans[name]; ok { // gm.GRTChannelMap.mutex.Lock()
							// time.Sleep(time.Millisecond * 200)
							gm.GRTChannelMap.mutex.Lock()
							grtchan.running = true
							gm.GRTChannelMap.mutex.Unlock()
							gm.wg.Add(1)
							fc.(func())()
							gm.wg.Done()
							gm.GRTChannelMap.mutex.Lock()
							grtchan.running = false
							gm.GRTChannelMap.mutex.Unlock()
						} else {
							return
						}
					})
				}
			}

		}
	}(gm, name, start, fc, args...)
}

func (gm *GRManager) newloopTask(name string, start bool, fc interface{}, args ...interface{}) {
	go func(gm *GRManager, name string, start bool, fc interface{}, args ...interface{}) {
		// gm.GRTChannelMap.mutex.Lock()
		err := gm.GRTChannelMap.register(name)
		if err != nil {
			return
		}
		for {
			select {
			case info := <-gm.delall:
				taskInfo := strings.Split(info, ":")
				signal := taskInfo[0]
				if signal == "__D" {
					fmt.Println("Task [" + gm.GRTChannelMap.grtchans[name].name + "] quit")
					// gm.GRTChannelMap.mutex.Lock()
					// time.Sleep(time.Millisecond * 100)
					gm.GRTChannelMap.unregister(name)
					gm.delall <- DELETE
					return
				} else {
					fmt.Println("Unknown Signal")
				}

			case info := <-gm.GRTChannelMap.grtchans[name].msg:
				taskInfo := strings.Split(info, ":")
				signal, gid := taskInfo[0], taskInfo[1]
				if gid == strconv.Itoa(int(gm.GRTChannelMap.grtchans[name].gid)) {
					if signal == "__P" && gm.GRTChannelMap.grtchans[name].running{
						fmt.Println("Task [" + gm.GRTChannelMap.grtchans[name].name + "] stopped")
						start = false
					} else if signal == "__T" {
						fmt.Println("Task [" + gm.GRTChannelMap.grtchans[name].name + "] started")
						start = true
					} else if signal == "__D" {
						fmt.Println("Task [" + gm.GRTChannelMap.grtchans[name].name + "] quit")
						// gm.GRTChannelMap.mutex.Lock()
						gm.GRTChannelMap.unregister(name)
						return
					} 
					// else {
					// 	fmt.Println("Unknown Signal")
					// }
				}
			default:
				// fmt.Println("No Signal")
				time.Sleep(time.Millisecond * 500)
			}
			
			runn, oky := gm.GRTChannelMap.grtchans[name]
			if oky {
				if start && !runn.running {
					time.Sleep(time.Millisecond*500)
					if len(args) > 1 {
						_ = gm.pool.Submit(func() {
							if grtchan, ok := gm.GRTChannelMap.grtchans[name]; ok { // gm.GRTChannelMap.mutex.Lock()
								// time.Sleep(time.Millisecond * 200)
								gm.GRTChannelMap.mutex.Lock()
								grtchan.running = true
								gm.GRTChannelMap.mutex.Unlock()
								gm.wg.Add(1)
								fc.(func(...interface{}))(args)
								gm.wg.Done()
								gm.GRTChannelMap.mutex.Lock()
								grtchan.running = false
								gm.GRTChannelMap.mutex.Unlock()
							} else {
								return
							}
						})
					} else if len(args) == 1 {
						_ = gm.pool.Submit(func() {
							if grtchan, ok := gm.GRTChannelMap.grtchans[name]; ok {

								gm.GRTChannelMap.mutex.Lock()
								grtchan.running = true
								gm.GRTChannelMap.mutex.Unlock()
								gm.wg.Add(1)
								fc.(func(interface{}))(args[0])
								gm.wg.Done()
								gm.GRTChannelMap.mutex.Lock()
								grtchan.running = false
								gm.GRTChannelMap.mutex.Unlock()
							} else {
								return
							}
						})
					} else {
						_ = gm.pool.Submit(func() {
							if grtchan, ok := gm.GRTChannelMap.grtchans[name]; ok { // gm.GRTChannelMap.mutex.Lock()
								// time.Sleep(time.Millisecond * 200)
								gm.GRTChannelMap.mutex.Lock()
								grtchan.running = true
								gm.GRTChannelMap.mutex.Unlock()
								gm.wg.Add(1)
								fc.(func())()
								gm.wg.Done()
								gm.GRTChannelMap.mutex.Lock()
								grtchan.running = false
								gm.GRTChannelMap.mutex.Unlock()
							} else {
								return
							}
						})
					}
				}

			}

		}
	}(gm, name, start, fc, args...)
}
