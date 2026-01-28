package cassandra

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/gocql/gocql"

	"tidb-benchmarks/pkg/config"
)

type Client struct {
	session *gocql.Session
}

func Open(ctx context.Context, cfg config.Config) (*Client, error) {
	base := gocql.NewCluster(splitHosts(cfg.CassandraHosts)...)
	base.Timeout = 10 * time.Second
	base.ConnectTimeout = 10 * time.Second
	base.Consistency = parseConsistency(cfg.CassandraConsistency)
	if cfg.CassandraTLSSkipVerify {
		base.SslOpts = &gocql.SslOptions{Config: &tls.Config{InsecureSkipVerify: true}}
	}
	if cfg.CassandraUsername != "" {
		base.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.CassandraUsername,
			Password: cfg.CassandraPassword,
		}
	}

	// Bootstrap keyspace if it does not exist yet.
	bootstrapSess, err := base.CreateSession()
	if err != nil {
		return nil, err
	}
	defer bootstrapSess.Close()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	q := fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class':'SimpleStrategy','replication_factor':1}", cfg.CassandraKeyspace)
	if err := bootstrapSess.Query(q).WithContext(ctx).Exec(); err != nil {
		return nil, err
	}

	cluster := gocql.NewCluster(splitHosts(cfg.CassandraHosts)...)
	cluster.Keyspace = cfg.CassandraKeyspace
	cluster.Timeout = base.Timeout
	cluster.ConnectTimeout = base.ConnectTimeout
	cluster.Consistency = base.Consistency
	cluster.Authenticator = base.Authenticator
	cluster.SslOpts = base.SslOpts

	sess, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		sess.Close()
		return nil, ctx.Err()
	default:
	}

	return &Client{session: sess}, nil
}

func (c *Client) Name() string { return "cassandra" }

func (c *Client) PrepareSchema(ctx context.Context, cfg config.Config) error {
	q := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s.%s (
	id bigint,
	k bigint,
	c blob,
	PRIMARY KEY (id)
);
`, cfg.CassandraKeyspace, cfg.Table)
	return c.session.Query(q).WithContext(ctx).Exec()
}

func (c *Client) Truncate(ctx context.Context, cfg config.Config) error {
	q := fmt.Sprintf("TRUNCATE %s.%s", cfg.CassandraKeyspace, cfg.Table)
	return c.session.Query(q).WithContext(ctx).Exec()
}

func (c *Client) Insert(ctx context.Context, cfg config.Config, id int64, k int64, payload []byte) error {
	q := fmt.Sprintf("INSERT INTO %s.%s (id, k, c) VALUES (?, ?, ?)", cfg.CassandraKeyspace, cfg.Table)
	return c.session.Query(q, id, k, payload).WithContext(ctx).Exec()
}

func (c *Client) Read(ctx context.Context, cfg config.Config, id int64) ([]byte, error) {
	q := fmt.Sprintf("SELECT id, k, c FROM %s.%s WHERE id = ?", cfg.CassandraKeyspace, cfg.Table)
	var (
		ignoredID int64
		ignoredK  int64
		payload   []byte
	)
	if err := c.session.Query(q, id).WithContext(ctx).Consistency(parseConsistency(cfg.CassandraConsistency)).Scan(&ignoredID, &ignoredK, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (c *Client) Update(ctx context.Context, cfg config.Config, id int64, k int64, payload []byte) error {
	q := fmt.Sprintf("UPDATE %s.%s SET k = ?, c = ? WHERE id = ?", cfg.CassandraKeyspace, cfg.Table)
	return c.session.Query(q, k, payload, id).WithContext(ctx).Exec()
}

func (c *Client) Close() error {
	c.session.Close()
	return nil
}

func splitHosts(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return []string{"127.0.0.1"}
	}
	return out
}

func parseConsistency(s string) gocql.Consistency {
	s = strings.TrimSpace(strings.ToUpper(s))
	switch s {
	case "ONE":
		return gocql.One
	case "TWO":
		return gocql.Two
	case "THREE":
		return gocql.Three
	case "QUORUM":
		return gocql.Quorum
	case "LOCAL_QUORUM":
		return gocql.LocalQuorum
	case "ALL":
		return gocql.All
	default:
		return gocql.LocalQuorum
	}
}
