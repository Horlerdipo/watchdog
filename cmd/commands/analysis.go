package commands

import (
	"context"
	"fmt"
	"github.com/horlerdipo/watchdog/database"
	"github.com/horlerdipo/watchdog/enums"
	"log/slog"
	"time"
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
			aliases: []string{"an"},
			usage:   "Run analysis on any of the sites being monitored.",
			Log:     logger,
		},
	}
}

func Analysis(ctx context.Context, logger *slog.Logger, urlId int) {
	var recentDownTime time.Duration
	var lastCheckTime time.Duration
	timeNow := time.Now()

	fmt.Println("Running analysis...")
	db := InitiateDB(ctx, logger)
	defer db.Close()
	urlRepository := database.NewUrlRepository(db)
	urlStatusRepository := database.NewUrlStatusRepository(db)

	url, err := urlRepository.FindById(ctx, urlId)
	if err != nil {
		logger.Error("Unable to find site: "+err.Error(), urlId)
	}

	//check the last downtime and subtract if from now
	//and if there is no downtime, subtract it from the created at time
	recentDownTime = getRecentDownTime(ctx, &url, urlStatusRepository, logger)
	lastCheckStatus, err := urlStatusRepository.GetLastStatus(ctx, url.Id)
	if err != nil {
		logger.Error("Unable to fetch last check status: "+err.Error(), url.Id)
	}

	fmt.Printf("Site Status: %s\n", url.Status)
	switch url.Status {
	case enums.Healthy:
		fmt.Printf("Currently Up for: %v \n", recentDownTime)
	case enums.UnHealthy:
		fmt.Printf("Currently Down for: %v \n", recentDownTime)
	case enums.Pending:
		fmt.Println("No check has been performed yet.")
	}

	if lastCheckStatus.UrlId != 0 {
		lastCheckTime = timeNow.Sub(lastCheckStatus.Time)
		fmt.Printf("Last Checked: %v ago\n", lastCheckTime.Round(time.Second))
	}
}

func getRecentDownTime(ctx context.Context, url *database.Url, urlStatusRepository database.UrlStatusRepository, logger *slog.Logger) time.Duration {
	timeNow := time.Now()
	var recentDownTimeUrlStatus database.UrlStatus
	var err error

	if url.Status == enums.Healthy {
		recentDownTimeUrlStatus, err = urlStatusRepository.GetRecentStatus(ctx, url.Id, false)
	} else {
		recentDownTimeUrlStatus, err = urlStatusRepository.GetRecentStatus(ctx, url.Id, true)
	}
	if err != nil {
		logger.Error("Unable to fetch most recent downtime: "+err.Error(), url.Id)
	}

	if recentDownTimeUrlStatus.UrlId == 0 {
		logger.Info("No recent downtime found, calculating from creation time", url.Id)
		return timeNow.Sub(url.CreatedAt)
	} else {
		logger.Info("Recent downtime found, calculating from creation time", recentDownTimeUrlStatus.Status)
		return timeNow.Sub(recentDownTimeUrlStatus.Time)
	}
}
