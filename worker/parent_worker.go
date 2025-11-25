package worker

import (
	"context"
	"fmt"
	"github.com/horlerdipo/watchdog/core"
	"github.com/horlerdipo/watchdog/env"
	"github.com/horlerdipo/watchdog/supervisor"
	"github.com/redis/go-redis/v9"
	"log"
	"sync"
)

type ParentWorker struct {
	RedisClient              *redis.Client
	Signal                   chan bool
	Interval                 int
	Ctx                      context.Context
	WorkPool                 chan []string
	ChildWorkerPoolWaitGroup sync.WaitGroup
	Supervisor               *supervisor.Supervisor
}

func NewParentWorker(ctx context.Context, redisClient *redis.Client, interval int, supervisor *supervisor.Supervisor) *ParentWorker {
	return &ParentWorker{
		Ctx:                      ctx,
		RedisClient:              redisClient,
		Interval:                 interval,
		Signal:                   make(chan bool),
		WorkPool:                 make(chan []string),
		ChildWorkerPoolWaitGroup: sync.WaitGroup{},
		Supervisor:               supervisor,
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
	urlLength, err := pw.RedisClient.LLen(pw.Ctx, core.FormatRedisList(pw.Interval)).Result()
	if err != nil {
		log.Println(err)
		return
	}

	if urlLength < 1 {
		return
	}

	//fetch all the ids in Redis
	urlIds, err := pw.RedisClient.LRange(pw.Ctx, core.FormatRedisList(pw.Interval), 0, urlLength).Result()
	fmt.Println(urlIds)
	if err != nil {
		log.Println(err)
		return
	}
	maxPoolSize := env.FetchInt("MAXIMUM_WORK_POOL_SIZE")
	if len(urlIds) <= maxPoolSize {
		pw.WorkPool <- urlIds
		return
	}
	pw.WorkPool <- urlIds
	return
}

func (pw *ParentWorker) spawnChildWorkers(maxChildWorkers int) {
	for i := 0; i < maxChildWorkers; i++ {
		pw.ChildWorkerPoolWaitGroup.Add(1)
		child := NewChildWorker(
			pw.Ctx,
			i+1,
			pw,
		)
		go child.Start()
	}
}
