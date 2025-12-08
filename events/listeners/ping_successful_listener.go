package listeners

import (
	"context"
	"fmt"
	"github.com/horlerdipo/watchdog/core"
	"github.com/horlerdipo/watchdog/database"
	"github.com/horlerdipo/watchdog/enums"
	"github.com/horlerdipo/watchdog/events"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

type PingSuccessfulListener struct {
	ctx context.Context
	log *slog.Logger
	DB  *pgxpool.Pool
}

func (sl *PingSuccessfulListener) Handle(event core.Event) {
	e := event.(*events.PingSuccessful)
	fmt.Printf("%v is healthy, pushing to timescale DB \n", e.Url)
	urlStatusRepo := database.NewUrlStatusRepository(sl.DB)
	err := urlStatusRepo.Add(sl.ctx, e.UrlId, e.Healthy)
	if err != nil {
		sl.log.Error(err.Error(), e)
		return
	}

	siteHealth := enums.Healthy
	if !e.Healthy {
		siteHealth = enums.UnHealthy
	}
	urlRepository := database.NewUrlRepository(sl.DB)
	err = urlRepository.UpdateStatus(sl.ctx, e.UrlId, siteHealth)
	if err != nil {
		sl.log.Error(err.Error(), e)
		return
	}
}

func NewPingSuccessfulListener(ctx context.Context, logger *slog.Logger, db *pgxpool.Pool) *PingSuccessfulListener {
	return &PingSuccessfulListener{
		log: logger,
		ctx: ctx,
		DB:  db,
	}
}
