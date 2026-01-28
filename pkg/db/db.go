package db

import (
	"context"
	"fmt"

	"tidb-benchmarks/pkg/config"
	"tidb-benchmarks/pkg/db/cassandra"
	"tidb-benchmarks/pkg/db/mysql"
)

type Client interface {
	Name() string
	PrepareSchema(ctx context.Context, cfg config.Config) error
	Truncate(ctx context.Context, cfg config.Config) error
	Insert(ctx context.Context, cfg config.Config, id int64, k int64, payload []byte) error
	Read(ctx context.Context, cfg config.Config, id int64) ([]byte, error)
	Update(ctx context.Context, cfg config.Config, id int64, k int64, payload []byte) error
	Close() error
}

func Open(ctx context.Context, cfg config.Config) (Client, error) {
	switch cfg.DB {
	case config.DBMySQL:
		return mysql.Open(ctx, cfg)
	case config.DBCassandra:
		return cassandra.Open(ctx, cfg)
	default:
		return nil, fmt.Errorf("unsupported db: %s", cfg.DB)
	}
}
