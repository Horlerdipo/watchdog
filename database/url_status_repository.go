package database

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UrlStatusRepository interface {
	Add(ctx context.Context, urlId int, status bool) error
}
type urlStatusRepository struct {
	pool *pgxpool.Pool
}

func (ur urlStatusRepository) Add(ctx context.Context, urlId int, status bool) error {
	sql := "INSERT INTO url_statuses (time, url_id,status) VALUES (NOW(), $1,$2)"

	_, err := ur.pool.Exec(ctx, sql, urlId, status)
	if err != nil {
		return err
	}
	return nil
}

func NewUrlStatusRepository(pool *pgxpool.Pool) UrlStatusRepository {
	return urlStatusRepository{
		pool: pool,
	}
}
