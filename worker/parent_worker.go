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
			i+1,
			&pw.ChildWorkerPoolMutex,
			&pw.ChildWorkerPoolWaitGroup,
			pw.WorkPool,
			pw.listName,
		)
		go child.Start()
	}
}
