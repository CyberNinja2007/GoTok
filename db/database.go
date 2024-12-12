package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Db *pgxpool.Pool
}

var (
	pgInstance *Postgres
	pgOnce     sync.Once
)

func NewPG(ctx context.Context, connString string) (*Postgres, error) {
	pgOnce.Do(func() {
		db, err := pgxpool.New(ctx, connString)
		if err != nil {
			pgInstance = nil
			err = fmt.Errorf("неудалось создать соединение с БД: %w", err)
			return
		}

		pgInstance = &Postgres{
			Db: db,
		}
	})

	return pgInstance, nil
}

func (pg *Postgres) Ping(ctx context.Context) error {
	return pg.Db.Ping(ctx)
}
