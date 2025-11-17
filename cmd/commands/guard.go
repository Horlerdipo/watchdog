package commands

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

type GuardCommand struct {
}

func (mc *GuardCommand) Arguments() []ArgumentContext {
	return []ArgumentContext{}
}

func (mc *GuardCommand) Name() string {
	return "guard"
}

func (mc *GuardCommand) Action(ctx context.Context, cmd CommandContext) error {
	Init(ctx)
	return nil
}

func (mc *GuardCommand) Aliases() []string {
	return []string{"g"}
}

func (mc *GuardCommand) Usage() string {
	return "Start the watchdog monitoring process."
}

func NewGuardCommand() *GuardCommand {
	return &GuardCommand{}
}

func Init(ctx context.Context) {
	redisClient := InitiateRedis(ctx)
	pool := InitiateDB(ctx)
	initiateOrchestrator(ctx, redisClient, pool)
	fmt.Println("Watchdog is running")
}

func InitiateRedis(ctx context.Context) *redis.Client {
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

	if err := redisClient.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("Redis connection failed: %v", err))
	}
	return redisClient
}

func InitiateDB(ctx context.Context) *pgxpool.Pool {
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

	newOrchestrator.AddIntervals(intervals)
	newOrchestrator.Start()
}
