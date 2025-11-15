package orchestrator

import (
	"context"
	"fmt"
	"github.com/horlerdipo/watchdog/supervisor"
	"github.com/horlerdipo/watchdog/worker"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

type Orchestrator struct {
	intervals   map[int]*worker.ParentWorker
	mutex       sync.RWMutex
	ctx         context.Context
	waitGroup   sync.WaitGroup
	RedisClient *redis.Client
	Supervisor  *supervisor.Supervisor
}

func NewOrchestrator(ctx context.Context, rdC *redis.Client) *Orchestrator {
	newSupervisor := supervisor.NewSupervisor(ctx)
	return &Orchestrator{
		intervals:   make(map[int]*worker.ParentWorker),
		ctx:         ctx,
		RedisClient: rdC,
		Supervisor:  newSupervisor,
	}
}

func (o *Orchestrator) Start() {
	fmt.Println("Orchestrator is running")
	for interval, parentWorker := range o.intervals {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		o.waitGroup.Add(1)
		go func() {
			for {
				select {
				case <-ticker.C:
					fmt.Printf("tick for %v \n", interval)
					parentWorker.Signal <- true
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

func (o *Orchestrator) FormatRedisList(interval int) string {
	return fmt.Sprintf("urls_to_monitor:%v", interval)
}

func (o *Orchestrator) AddInterval(interval int, worker *worker.ParentWorker) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.intervals[interval] = worker
}

func (o *Orchestrator) AddIntervals(intervals []int) {
	for _, interval := range intervals {
		workerGroup := worker.NewParentWorker(o.ctx, o.RedisClient, o.FormatRedisList(interval), o.Supervisor)
		workerGroup.Start()
		o.AddInterval(interval, workerGroup)
	}
}

func (o *Orchestrator) Intervals() []int {
	var intervals []int
	for interval, _ := range o.intervals {
		intervals = append(intervals, interval)
	}
	return intervals
}

func (o *Orchestrator) Stop() {}
