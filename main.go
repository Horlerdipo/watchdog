package main

import (
	"context"
	"fmt"
	"github.com/horlerdipo/watchdog/enums"
	"github.com/horlerdipo/watchdog/env"
	"github.com/horlerdipo/watchdog/orchestrator"
	"github.com/jackc/pgx/v5/pgxpool"
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

	pool := initiateDB(ctx)
	initiateOrchestrator(ctx, redisClient, pool)
	fmt.Println("Watchdog is running")
}

func initiateDB(ctx context.Context) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, fmt.Sprintf("postgres://%v:%v@%v:%v/%v", env.FetchString("DB_USER"), env.FetchString("DB_PASSWORD"), env.FetchString("DB_HOST"), env.FetchString("DB_PORT"), env.FetchString("DB_DATABASE")))
	if err != nil {
		panic(fmt.Sprintf("pgxpool connection failed: %v", err))
	}
	if err := pool.Ping(ctx); err != nil {
		panic(fmt.Sprintf("Unable to ping database: %v", err))
	}

	fmt.Println("Connected to PostgreSQL database!")
	return pool
}

func initiateOrchestrator(ctx context.Context, redisClient *redis.Client, pool *pgxpool.Pool) {
	newOrchestrator := orchestrator.NewOrchestrator(ctx, redisClient, pool)
	newOrchestrator.Supervisor.Activate()
	intervals := []enums.MonitoringFrequency{
		enums.TenSeconds, // 10 seconds
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
}
