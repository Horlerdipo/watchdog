package supervisor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

//This takes output from child workers as input
//Processes it for analytics with Timescale DB
//Dispatches Event for when a URL is unreachable

type Supervisor struct {
	WorkPool  chan Task
	BatchSize int
	Timeout   time.Duration
	ctx       context.Context
	WaitGroup *sync.WaitGroup
}

type Task struct {
	Healthy bool
	Url     string
}

func NewSupervisor(ctx context.Context, batchSize int, Timeout time.Duration) *Supervisor {
	return &Supervisor{
		WorkPool:  make(chan Task, batchSize),
		ctx:       ctx,
		BatchSize: batchSize,
		Timeout:   Timeout,
		WaitGroup: &sync.WaitGroup{},
	}
}

func (s *Supervisor) Activate() {
	buffer := make([]Task, 0, s.BatchSize)
	ticker := time.NewTicker(s.Timeout)

	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case task := <-s.WorkPool:
				buffer = append(buffer, task)
				if len(buffer) >= s.BatchSize {
					s.flush(buffer)
					buffer = buffer[:0]
				}
			case <-ticker.C:
				if len(buffer) > 0 {
					s.flush(buffer)
					buffer = buffer[:0]
				}
			}
		}
	}()
}

func (s *Supervisor) flush(buffer []Task) {
	for _, task := range buffer {
		fmt.Printf("supervisor picked up new task %v\n", task.Url)
		if task.Healthy {
			fmt.Printf("%v is healthy, pushing to timescale DB \n", task.Url)
		} else {
			fmt.Printf("%v is unhealthy, pushing to timescale DB and dispatching downtime notification event \n", task.Url)
		}
	}
}
