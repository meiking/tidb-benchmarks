package mysql

import (
	"context"
	"database/sql"
	"fmt"

	mysqlDriver "github.com/go-sql-driver/mysql"

	"tidb-benchmarks/pkg/config"
)

type Client struct {
	db *sql.DB
}

func Open(ctx context.Context, cfg config.Config) (*Client, error) {
	dsn := cfg.MySQLDSN
	parsed, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}

	if cfg.MySQLTLS {
		// Enable TLS by default unless DSN already specifies tls.
		if parsed.TLSConfig == "" {
			parsed.TLSConfig = "true"
		}
	} else {
		// Explicitly disable TLS (even if DSN had tls=...).
		parsed.TLSConfig = "false"
	}

	dsn = parsed.FormatDSN()

	dbConn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := dbConn.PingContext(ctx); err != nil {
		_ = dbConn.Close()
		return nil, err
	}

	dbConn.SetMaxOpenConns(cfg.Threads * 4)
	dbConn.SetMaxIdleConns(cfg.Threads * 2)

	return &Client{db: dbConn}, nil
}

func (c *Client) Name() string { return "mysql" }

func (c *Client) PrepareSchema(ctx context.Context, cfg config.Config) error {
	ddl := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	id BIGINT NOT NULL,
	k BIGINT NOT NULL,
	c BLOB NOT NULL,
	PRIMARY KEY (id)
) ENGINE=InnoDB;
`, cfg.Table)
	_, err := c.db.ExecContext(ctx, ddl)
	return err
}

func (c *Client) Truncate(ctx context.Context, cfg config.Config) error {
	_, err := c.db.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s", cfg.Table))
	return err
}

func (c *Client) Insert(ctx context.Context, cfg config.Config, id int64, k int64, payload []byte) error {
	_, err := c.db.ExecContext(ctx, fmt.Sprintf("INSERT INTO %s (id, k, c) VALUES (?, ?, ?)", cfg.Table), id, k, payload)
	return err
}

func (c *Client) Read(ctx context.Context, cfg config.Config, id int64) ([]byte, error) {
	row := c.db.QueryRowContext(ctx, fmt.Sprintf("SELECT id, k, c FROM %s WHERE id = ?", cfg.Table), id)
	var (
		ignoredID int64
		ignoredK  int64
		payload   []byte
	)
	if err := row.Scan(&ignoredID, &ignoredK, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *Client) Update(ctx context.Context, cfg config.Config, id int64, k int64, payload []byte) error {
	_, err := c.db.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET k = ?, c = ? WHERE id = ?", cfg.Table), k, payload, id)
	return err
}

func (c *Client) Close() error { return c.db.Close() }
