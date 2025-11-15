package worker

import (
	"context"
	"fmt"
	"github.com/horlerdipo/watchdog/env"
	"github.com/horlerdipo/watchdog/supervisor"
	"net/http"
	"time"
)

type ChildWorker struct {
	Id           int
	Ctx          context.Context
	ParentWorker *ParentWorker
}

func NewChildWorker(ctx context.Context, id int, parentWorker *ParentWorker) *ChildWorker {
	return &ChildWorker{
		Id:           id,
		Ctx:          ctx,
		ParentWorker: parentWorker,
	}
}

func (cw *ChildWorker) Start() {
	defer cw.ParentWorker.ChildWorkerPoolWaitGroup.Done()
	fmt.Printf("Child Worker %d with parent %v started and waiting for tasks\n", cw.Id, cw.ParentWorker.listName)
	for {
		select {
		case <-cw.Ctx.Done():
			fmt.Printf("Child Worker %d with parent %v shutting down\n", cw.Id, cw.ParentWorker.listName)
			return

		case urls, ok := <-cw.ParentWorker.WorkPool:
			if !ok {
				fmt.Printf("Worker %d: WorkPool with parent %v closed, shutting down\n", cw.Id, cw.ParentWorker.listName)
				return
			}
			fmt.Printf("Worker %d with parent %v picked up chunk of %d URLs\n", cw.Id, cw.ParentWorker.listName, len(urls))

			for _, url := range urls {
				fmt.Printf("Worker %d with parent %v processing: %s\n", cw.Id, cw.ParentWorker.listName, url)
				cw.Work(url)
			}
			fmt.Printf("Worker %d with parent %v completed chunk\n", cw.Id, cw.ParentWorker.listName)
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
	fmt.Printf("Worker %d with parent %v tried monitoring %v and returned %v \n", cw.Id, cw.ParentWorker.listName, url, resp.StatusCode)
	task := supervisor.Task{
		Healthy: resp.StatusCode > 199 && resp.StatusCode < 300,
		Url:     url,
	}

	cw.ParentWorker.Supervisor.WorkPool <- task
	return
}
