package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitDB(ctx context.Context, dbUrl string) error {
	cfg, err := pgxpool.ParseConfig(dbUrl)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse db url: %v\n", err)
		os.Exit(1)
	}

	cfg.MaxConns = 10

	pool, err := pgxpool.NewWithConfig(ctx, cfg)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create a new pool: %v\n", err)
		os.Exit(1)
	}

	newCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

	defer cancel()

	e := pool.Ping(newCtx)

	if e != nil {
		pool.Close()

		fmt.Fprintf(os.Stderr, "Unable to ping to database: %v\n", e)
		os.Exit(1)
	}

	DB = pool

	return nil
}
