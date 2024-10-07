package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	query = `
insert into hashes(hash) values ($1);
`
	uniqueViolationPgCode = "23505"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) InsertHash(ctx context.Context, hash []byte) (bool, error) {
	_, err := r.pool.Exec(ctx, query, hash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationPgCode {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
