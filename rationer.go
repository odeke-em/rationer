package rationer

import (
	"sync"
	"time"
)

var Sentinel = "sentinel"

type Job func() interface{}

type Rationer struct {
	finished bool
	dibbs    chan interface{}
	results  chan interface{}
	ready    chan bool
	done     chan bool
}

func NewRationer(capacity uint64) *Rationer {
	ready := make(chan bool)
	dibbs := make(chan interface{}, capacity)
	results := make(chan interface{})
	done := make(chan bool, 1)

	r := Rationer{
		dibbs:    dibbs,
		ready:    ready,
		done:     done,
		results:  results,
		finished: false,
	}

	return &r
}

func (r *Rationer) readyTick() {
	var mu sync.Mutex

	tick := time.Tick(1e9)
	for !r.finished {
		mu.Lock()
		if len(r.dibbs) < cap(r.dibbs) {
			r.ready <- true
		}
		mu.Unlock()
		<-tick
	}
}

func (r *Rationer) Run() chan interface{} {
	loader := make(chan interface{})

	go func() {
		// Spin while waiting till ready
		spinning := true
		for spinning {
			select {
			case <-r.ready:
				for head := range loader {
					if head == Sentinel { // Cancellation requested
						spinning = false
						break
					}

					job, ok := head.(Job)
					if !ok {
						continue
					}

					go func() {
						r.dibbs <- true
						result := job()

						mu := sync.Mutex{}
						mu.Lock()
						if len(r.dibbs) >= 1 {
							<-r.dibbs
						}
						mu.Unlock()

						r.results <- result
					}()

					break
				}
			default:
			}
		}

		r.done <- true
	}()

	go r.readyTick()

	return loader
}

func (r *Rationer) Wait() {
	r.finished = true
	<-r.done
}

func (r *Rationer) Results() chan interface{} {
	return r.results
}
