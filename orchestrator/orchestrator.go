package orchestrator

import (
	"context"
	"fmt"
	"github.com/horlerdipo/watchdog/database"
	"github.com/horlerdipo/watchdog/enums"
	"github.com/horlerdipo/watchdog/env"
	"github.com/horlerdipo/watchdog/supervisor"
	"github.com/horlerdipo/watchdog/worker"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

type Orchestrator struct {
	intervals     map[int]*worker.ParentWorker
	mutex         sync.RWMutex
	ctx           context.Context
	waitGroup     sync.WaitGroup
	RedisClient   *redis.Client
	Supervisor    *supervisor.Supervisor
	DB            *pgxpool.Pool
	UrlRepository database.UrlRepository
}

func NewOrchestrator(ctx context.Context, rdC *redis.Client, pool *pgxpool.Pool) *Orchestrator {
	newSupervisor := supervisor.NewSupervisor(
		ctx,
		env.FetchInt("SUPERVISOR_POOL_FLUSH_BATCHSIZE", 100),
		time.Duration(env.FetchInt("SUPERVISOR_POOL_FLUSH_TIMEOUT", 5))*time.Second,
	)

	return &Orchestrator{
		intervals:     make(map[int]*worker.ParentWorker),
		ctx:           ctx,
		RedisClient:   rdC,
		Supervisor:    newSupervisor,
		DB:            pool,
		UrlRepository: database.NewUrlRepository(pool),
	}
}

func (o *Orchestrator) Start() {
	fmt.Println("Orchestrator is running")
	o.PrefillRedisList(o.ctx)
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

func (o *Orchestrator) AddInterval(interval enums.MonitoringFrequency, worker *worker.ParentWorker) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.intervals[interval.ToSeconds()] = worker
}

func (o *Orchestrator) AddIntervals(intervals []enums.MonitoringFrequency) {
	for _, interval := range intervals {
		workerGroup := worker.NewParentWorker(o.ctx, o.RedisClient, o.FormatRedisList(interval.ToSeconds()), o.Supervisor)
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

func (o *Orchestrator) PrefillRedisList(ctx context.Context) {
	urls, err := database.NewUrlRepository(o.DB).FetchAll(ctx, 10, 0)
	if err != nil {
		panic(err)
	}

	for _, interval := range o.Intervals() {
		o.RedisClient.Del(ctx, o.FormatRedisList(interval))
	}

	for _, url := range urls {
		o.RedisClient.LPush(ctx, o.FormatRedisList(url.MonitoringFrequency.ToSeconds()), url.Url)
	}
}
