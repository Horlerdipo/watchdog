package worker

import (
	"context"
	"fmt"
	"github.com/horlerdipo/watchdog/env"
	"github.com/redis/go-redis/v9"
	"log"
	"sync"
)

type ParentWorker struct {
	RedisClient              *redis.Client
	Signal                   chan bool
	listName                 string
	Ctx                      context.Context
	WorkPool                 chan []string
	ChildWorkerPoolMutex     sync.Mutex
	ChildWorkerPoolWaitGroup sync.WaitGroup
}

func NewParentWorker(ctx context.Context, redisClient *redis.Client, listName string) *ParentWorker {
	return &ParentWorker{
		Ctx:                      ctx,
		RedisClient:              redisClient,
		listName:                 listName,
		Signal:                   make(chan bool),
		WorkPool:                 make(chan []string),
		ChildWorkerPoolMutex:     sync.Mutex{},
		ChildWorkerPoolWaitGroup: sync.WaitGroup{},
	}
}

func (pw *ParentWorker) Start() {
	maxChildWorkers := env.FetchInt("MAXIMUM_CHILD_WORKERS")
	if maxChildWorkers < 1 {
		panic("MAXIMUM_CHILD_WORKERS must be greater than 0")
	}
	pw.spawnChildWorkers(maxChildWorkers)
	go func() {
		for {
			select {
			case <-pw.Ctx.Done():
				return
			case <-pw.Signal:
				pw.Work()
			}
		}
	}()
}

func (pw *ParentWorker) Work() {
	fmt.Println("Parent worker is running")
	urlLength, err := pw.RedisClient.LLen(pw.Ctx, pw.listName).Result()
	if err != nil {
		log.Println(err)
		return
	}

	if urlLength < 1 {
		return
	}

	//fetch all the url in Redis
	urls, err := pw.RedisClient.LRange(pw.Ctx, pw.listName, 0, urlLength).Result()
	fmt.Println(urls)
	if err != nil {
		log.Println(err)
		return
	}
	maxPoolSize := env.FetchInt("MAXIMUM_WORK_POOL_SIZE")
	if len(urls) <= maxPoolSize {
		pw.WorkPool <- urls
		return
	}
	pw.WorkPool <- urls
	return
}

func (pw *ParentWorker) spawnChildWorkers(maxChildWorkers int) {
	for i := 0; i < maxChildWorkers; i++ {
		pw.ChildWorkerPoolWaitGroup.Add(1)
		child := NewChildWorker(
			pw.Ctx,
			&pw.ChildWorkerPoolMutex,
			&pw.ChildWorkerPoolWaitGroup,
			pw.WorkPool,
		)
		go child.Start(i)
	}
}

type ChildWorker struct {
	Ctx       context.Context
	mutex     *sync.Mutex
	waitGroup *sync.WaitGroup
	WorkPool  chan []string
}

func NewChildWorker(ctx context.Context, mutex *sync.Mutex, waitGroup *sync.WaitGroup, workPool chan []string) *ChildWorker {
	return &ChildWorker{
		Ctx:       ctx,
		mutex:     mutex,
		waitGroup: waitGroup,
		WorkPool:  workPool,
	}
}

func (cw *ChildWorker) Start(hierarchy int) {
	defer cw.waitGroup.Done()
	fmt.Printf("Child Worker %d started and waiting for tasks\n", hierarchy)
	for {
		select {
		case <-cw.Ctx.Done():
			fmt.Printf("Child Worker %d shutting down\n", hierarchy)
			return

		case urls, ok := <-cw.WorkPool:
			if !ok {
				fmt.Printf("Worker %d: WorkPool closed, shutting down\n", hierarchy)
				return
			}
			fmt.Printf("Worker %d picked up chunk of %d URLs\n", hierarchy, len(urls))

			for _, url := range urls {
				fmt.Printf("Worker %d processing: %s\n", hierarchy, url)
			}

			fmt.Printf("Worker %d completed chunk\n", hierarchy)
		}
	}
}
