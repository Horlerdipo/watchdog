package commands

import (
	"context"
	"fmt"
	"github.com/horlerdipo/watchdog/database"
	"github.com/horlerdipo/watchdog/enums"
	"log/slog"
)

type AnalysisCommand struct {
	*BaseCommand
}

func (mc *AnalysisCommand) Action(ctx context.Context, cmd CommandContext) error {
	urlId := cmd.Int("id")
	Analysis(ctx, mc.Log, urlId)
	return nil
}

func (mc *AnalysisCommand) Arguments() []ArgumentContext {
	return []ArgumentContext{
		{
			Name:    "id",
			Usage:   "The ID of the site you want analysis on.",
			Type:    enums.Int,
			Default: 0,
		},
	}
}

func NewAnalysisCommand(logger *slog.Logger) *AnalysisCommand {
	return &AnalysisCommand{
		BaseCommand: &BaseCommand{
			name:    "analysis",
			aliases: []string{"a"},
			usage:   "Run analysis on any of the sites being monitored.",
			Log:     logger,
		},
	}
}

func Analysis(ctx context.Context, logger *slog.Logger, urlId int) {
	fmt.Println("Running analysis...")
	db := InitiateDB(ctx, logger)
	defer db.Close()
	urlRepository := database.NewUrlRepository(db)
	url, err := urlRepository.FindById(ctx, urlId)
	if err != nil {
		logger.Error("Unable to find site: "+err.Error(), urlId)
	}
	fmt.Printf("Site Status: %s\n", url.Status)
}
