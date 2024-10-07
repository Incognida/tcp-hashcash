package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	hasIpQuery = `
select exists(select 1 from ips where ip = $1);
`
	insertIpQuery = `
insert into ips(ip) values ($1) on conflict(ip) do nothing;
`
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) HasIP(ctx context.Context, ip string) (bool, error) {
	var exists bool

	err := r.pool.QueryRow(ctx, hasIpQuery, ip).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *Repository) InsertIP(ctx context.Context, ip string) error {
	_, err := r.pool.Exec(ctx, insertIpQuery, ip)

	return err
}
