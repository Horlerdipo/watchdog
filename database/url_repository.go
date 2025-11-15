package database

import (
	"context"
	"fmt"
	"github.com/horlerdipo/watchdog/enums"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UrlRepository interface {
	FetchAll(ctx context.Context, limit int, offset int) ([]Url, error)
}
type urlRepository struct {
	pool *pgxpool.Pool
}

func (ur urlRepository) FetchAll(ctx context.Context, limit int, offset int) ([]Url, error) {
	sql := "SELECT id,url,http_method,status,monitoring_frequency,created_at,updated_at FROM urls"
	rows, err := ur.pool.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []Url
	var monitoringFrequency string
	var status string
	var httpMethod string

	for rows.Next() {
		var url Url
		err := rows.Scan(
			&url.Id,
			&url.Url,
			&httpMethod,
			&status,
			&monitoringFrequency,
			&url.CreatedAt,
			&url.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		parsedMonitoringFrequency, err := enums.ParseMonitoringFrequency(monitoringFrequency)
		if err != nil {
			return nil, err
		}

		parsedStatus, err := enums.ParseSiteHealth(status)
		if err != nil {
			return nil, err
		}

		parsedHttpMethod, err := enums.ParseHttpMethod(httpMethod)
		if err != nil {
			return nil, err
		}

		url.MonitoringFrequency = parsedMonitoringFrequency
		url.Status = parsedStatus
		url.HttpMethod = parsedHttpMethod
		urls = append(urls, url)
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating task rows: %w", err)
		}
	}
	return urls, nil
}

func NewUrlRepository(pool *pgxpool.Pool) UrlRepository {
	return urlRepository{
		pool: pool,
	}
}
