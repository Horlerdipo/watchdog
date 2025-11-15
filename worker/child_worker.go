package worker

import (
	"context"
	"fmt"
	"github.com/horlerdipo/watchdog/env"
	"net/http"
	"sync"
	"time"
)

type ChildWorker struct {
	Id           int
	Ctx          context.Context
	mutex        *sync.Mutex
	waitGroup    *sync.WaitGroup
	WorkPool     chan []string
	ParentWorker string
}

func NewChildWorker(ctx context.Context, id int, mutex *sync.Mutex, waitGroup *sync.WaitGroup, workPool chan []string, parentWorker string) *ChildWorker {
	return &ChildWorker{
		Id:           id,
		Ctx:          ctx,
		mutex:        mutex,
		waitGroup:    waitGroup,
		WorkPool:     workPool,
		ParentWorker: parentWorker,
	}
}

func (cw *ChildWorker) Start() {
	defer cw.waitGroup.Done()
	fmt.Printf("Child Worker %d with parent %v started and waiting for tasks\n", cw.Id, cw.ParentWorker)
	for {
		select {
		case <-cw.Ctx.Done():
			fmt.Printf("Child Worker %d with parent %v shutting down\n", cw.Id, cw.ParentWorker)
			return

		case urls, ok := <-cw.WorkPool:
			if !ok {
				fmt.Printf("Worker %d: WorkPool with parent %v closed, shutting down\n", cw.Id, cw.ParentWorker)
				return
			}
			fmt.Printf("Worker %d with parent %v picked up chunk of %d URLs\n", cw.Id, cw.ParentWorker, len(urls))

			for _, url := range urls {
				fmt.Printf("Worker %d with parent %v processing: %s\n", cw.Id, cw.ParentWorker, url)
				cw.Work(url)
			}
			fmt.Printf("Worker %d with parent %v completed chunk\n", cw.Id, cw.ParentWorker)
		}
	}
}

func (cw *ChildWorker) Work(url string) {
	client := &http.Client{
		Timeout: time.Duration(env.FetchInt("REQUEST_TIMEOUT", 5)) * time.Second,
	}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("Worker %d with parent %v tried monitoring %v and returned %v \n", cw.Id, cw.ParentWorker, url, resp.StatusCode)
}
