package main

import (
	"context"
	"fmt"
	"github.com/horlerdipo/watchdog/orchestrator"
	redis2 "github.com/redis/go-redis/v9"
)

func main() {
	fmt.Println("Watchdog is running")
	redis := redis2.NewClient(&redis2.Options{})
	ctx := context.Background()
	ctx = context.WithoutCancel(ctx)

	newOrchestrator := orchestrator.NewOrchestrator(ctx)
	intervals := []int{
		10,    // 10 seconds
		30,    // 30 seconds
		60,    // 1 minute
		300,   // 5 minutes
		1800,  // 30 minutes
		3600,  // 1 hour
		43200, // 12 hour
		86400, // 24 hour
	}
	newOrchestrator.AddIntervals(intervals)
	fmt.Println(newOrchestrator.Intervals())
	newOrchestrator.Start()
}
