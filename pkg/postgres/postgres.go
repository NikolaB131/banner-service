package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func New(url string) (*Postgres, error) {
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Unable to create postgres pool: %s", err.Error()))
	}
	err = dbpool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("Ping to postgres pool failed: %s", err.Error()))
	}

	return &Postgres{Pool: dbpool}, nil
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
