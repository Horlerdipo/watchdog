package main

import (
	"context"
	"fmt"
	"github.com/horlerdipo/watchdog/env"
	"github.com/horlerdipo/watchdog/orchestrator"
	"github.com/redis/go-redis/v9"
	"time"
)

func main() {
	env.LoadEnv(".env")
	redisClient := redis.NewClient(&redis.Options{
		Addr:         env.FetchString("REDIS_HOST"),
		DB:           env.FetchInt("REDIS_DB"),
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
	})
	ctx := context.Background()
	ctx = context.WithoutCancel(ctx)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("Redis connection failed: %v", err))
	}

	redisClient.Set(ctx, "key", "value", 0)
	newOrchestrator := orchestrator.NewOrchestrator(ctx, redisClient)
	newOrchestrator.Supervisor.Activate()
	intervals := []int{
		10, // 10 seconds
		//30,    // 30 seconds
		//60,    // 1 minute
		//300,   // 5 minutes
		//1800,  // 30 minutes
		//3600,  // 1 hour
		//43200, // 12 hours
		//86400, // 24 hours
	}

	//for _, interval := range intervals {
	//	redisClient.LPush(ctx, fmt.Sprintf("urls_to_monitor:%v", interval), "https://google.com")
	//}

	newOrchestrator.AddIntervals(intervals)
	fmt.Println(newOrchestrator.Intervals())
	newOrchestrator.Start()
	fmt.Println("Watchdog is running")
}
