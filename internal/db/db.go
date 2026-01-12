package db


import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Addr        string
	MaxConns    int32
	MinConns    int32
	MaxIdleTime string
}

func New(config Config) (*pgxpool.Pool, error) {

	cfg, err := pgxpool.ParseConfig(config.Addr)
	if err != nil {
		return nil, err
	}

	// connection pool settings
	cfg.MaxConns = config.MaxConns
	cfg.MinConns = config.MinConns

	if config.MaxIdleTime != "" {
		duration, err := time.ParseDuration(config.MaxIdleTime)
		if err != nil {
			return nil, err
		}
		cfg.MaxConnIdleTime = duration
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// check connectivity
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}
