package worker

import (
	"context"
	"fmt"
	"sync"
)

type ChildWorker struct {
	Ctx          context.Context
	mutex        *sync.Mutex
	waitGroup    *sync.WaitGroup
	WorkPool     chan []string
	ParentWorker string
}

func NewChildWorker(ctx context.Context, mutex *sync.Mutex, waitGroup *sync.WaitGroup, workPool chan []string, parentWorker string) *ChildWorker {
	return &ChildWorker{
		Ctx:          ctx,
		mutex:        mutex,
		waitGroup:    waitGroup,
		WorkPool:     workPool,
		ParentWorker: parentWorker,
	}
}

func (cw *ChildWorker) Start(hierarchy int) {
	defer cw.waitGroup.Done()
	fmt.Printf("Child Worker %d with parent %v started and waiting for tasks\n", hierarchy, cw.ParentWorker)
	for {
		select {
		case <-cw.Ctx.Done():
			fmt.Printf("Child Worker %d with parent %v shutting down\n", hierarchy, cw.ParentWorker)
			return

		case urls, ok := <-cw.WorkPool:
			if !ok {
				fmt.Printf("Worker %d: WorkPool with parent %v closed, shutting down\n", hierarchy, cw.ParentWorker)
				return
			}
			fmt.Printf("Worker %d with parent %v picked up chunk of %d URLs\n", hierarchy, cw.ParentWorker, len(urls))

			for _, url := range urls {
				fmt.Printf("Worker %d with parent %v processing: %s\n", hierarchy, cw.ParentWorker, url)
			}

			fmt.Printf("Worker %d with parent %v completed chunk\n", hierarchy, cw.ParentWorker)
		}
	}
}
