package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Orchestrator struct {
	intervals []int
	mutex     sync.RWMutex
	ctx       context.Context
	waitGroup sync.WaitGroup
}

func NewOrchestrator(ctx context.Context) *Orchestrator {
	return &Orchestrator{
		ctx: ctx,
	}
}

func (o *Orchestrator) Start() {
	fmt.Println("Orchestrator is running")
	for _, interval := range o.intervals {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		o.waitGroup.Add(1)
		go func() {
			for {
				select {
				case <-ticker.C:
					fmt.Printf("tick for %v \n", interval)
				case <-o.ctx.Done():
					fmt.Println("Orchestrator is stopped")
					ticker.Stop()
					o.waitGroup.Done()
					return
				}
			}
		}()
	}
	o.waitGroup.Wait()
}

func (o *Orchestrator) AddInterval(interval int) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.intervals = append(o.intervals, interval)
}

func (o *Orchestrator) AddIntervals(intervals []int) {
	for _, interval := range intervals {
		o.AddInterval(interval)
	}
}

func (o *Orchestrator) Intervals() []int {
	return o.intervals
}

func (o *Orchestrator) Stop() {}
