package commands

import (
	"context"
	"fmt"
	"github.com/horlerdipo/watchdog/database"
	"github.com/horlerdipo/watchdog/enums"
)

type RemoveCommand struct {
}

func (mc *RemoveCommand) Name() string {
	return "remove"
}

func (mc *RemoveCommand) Aliases() []string {
	return []string{"r"}
}

func (mc *RemoveCommand) Usage() string {
	return "Remove a URL from the watchdog monitoring process."
}

func (mc *RemoveCommand) Arguments() []ArgumentContext {
	return []ArgumentContext{
		{
			Name:    "id",
			Usage:   "The ID of the URL to be removed.",
			Type:    enums.Int,
			Default: 0,
		},
	}
}

func (mc *RemoveCommand) Action(ctx context.Context, cmd CommandContext) error {
	id := cmd.Int("id")

	// Check if required argument is provided
	if id == 0 {
		return fmt.Errorf("ID is required")
	}

	pool := InitiateDB(ctx)
	urlRepository := database.NewUrlRepository(pool)

	url, err := urlRepository.FindById(ctx, id)
	if err != nil {
		fmt.Printf("Error removing url: %v", err)
		return err
	}

	err = urlRepository.Delete(ctx, url.Id)

	if err != nil {
		fmt.Printf("Error removing url: %v", err)
		return err
	}

	redisClient := InitiateRedis(ctx)
	err = RefreshRedisInterval(ctx, redisClient, pool, url.MonitoringFrequency)
	if err != nil {
		fmt.Printf("Error removing url from redis: %v", err)
		return err
	}

	fmt.Printf("URL successfully removing, ID: %v", id)
	return nil
}

func NewRemoveCommand() *RemoveCommand {
	return &RemoveCommand{}
}
