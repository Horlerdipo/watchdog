package supervisor

import (
	"context"
	"fmt"
	"sync"
)

//This takes output from child workers as input
//Processes it for analytics with Timescale DB
//Dispatches Event for when a URL is unreachable

type Supervisor struct {
	WorkPool  chan Task
	ctx       context.Context
	WaitGroup *sync.WaitGroup
}

type Task struct {
	Healthy bool
	Url     string
}

func NewSupervisor(ctx context.Context) *Supervisor {
	return &Supervisor{
		WorkPool:  make(chan Task),
		ctx:       ctx,
		WaitGroup: &sync.WaitGroup{},
	}
}

func (s *Supervisor) Activate() {
	//s.WaitGroup.Add(1)
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case task := <-s.WorkPool:
				fmt.Printf("supervisor picked up new task %v", task.Url)
				if task.Healthy {
					fmt.Printf("%v is healthy, pushing to timescale DB", task.Url)
				} else {
					fmt.Printf("%v is unhealthy, pushing to timescale DB and dispatching downtime notification event", task.Url)
				}
			}
		}
	}()

}
