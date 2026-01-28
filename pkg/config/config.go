package config

import (
	"time"

	"github.com/spf13/pflag"
)

type DBKind string

const (
	DBMySQL     DBKind = "mysql"
	DBCassandra DBKind = "cassandra"
)

type OutputFormat string

const (
	OutputText OutputFormat = "text"
	OutputJSON OutputFormat = "json"
)

type Config struct {
	DB DBKind

	MySQLDSN           string
	MySQLTLSSkipVerify bool

	CassandraHosts         string
	CassandraKeyspace      string
	CassandraUsername      string
	CassandraPassword      string
	CassandraConsistency   string
	CassandraTLSSkipVerify bool

	Table       string
	TableSize   int64
	PayloadSize int

	Threads int
	Time    time.Duration
	Timeout time.Duration

	ReadRatio float64

	Warmup time.Duration

	Output OutputFormat
}

func Default() Config {
	return Config{
		DB:                   DBMySQL,
		MySQLDSN:             "root:@tcp(127.0.0.1:3306)/test?parseTime=true&multiStatements=true",
		CassandraHosts:       "127.0.0.1",
		CassandraKeyspace:    "bench",
		CassandraConsistency: "LOCAL_QUORUM",
		Table:                "sbtest",
		TableSize:            100000,
		PayloadSize:          120,
		Threads:              16,
		Time:                 30 * time.Second,
		Timeout:              10 * time.Minute,
		ReadRatio:            0.5,
		Warmup:               2 * time.Second,
		Output:               OutputText,
	}
}

func BindCommonFlags(fs *pflag.FlagSet, cfg *Config) {
	fs.StringVar((*string)(&cfg.DB), "db", string(cfg.DB), "Target database: mysql|cassandra")
	fs.StringVar(&cfg.MySQLDSN, "mysql-dsn", cfg.MySQLDSN, "MySQL DSN")
	fs.BoolVar(&cfg.MySQLTLSSkipVerify, "mysql-tls-skip-verify", cfg.MySQLTLSSkipVerify, "Skip TLS certificate/hostname verification for MySQL (INSECURE)")
	fs.StringVar(&cfg.CassandraHosts, "cassandra-hosts", cfg.CassandraHosts, "Cassandra hosts, comma-separated")
	fs.StringVar(&cfg.CassandraKeyspace, "cassandra-keyspace", cfg.CassandraKeyspace, "Cassandra keyspace")
	fs.StringVar(&cfg.CassandraUsername, "cassandra-username", cfg.CassandraUsername, "Cassandra username")
	fs.StringVar(&cfg.CassandraPassword, "cassandra-password", cfg.CassandraPassword, "Cassandra password")
	fs.StringVar(&cfg.CassandraConsistency, "cassandra-consistency", cfg.CassandraConsistency, "Cassandra consistency (e.g. ONE, LOCAL_QUORUM)")
	fs.BoolVar(&cfg.CassandraTLSSkipVerify, "cassandra-tls-skip-verify", cfg.CassandraTLSSkipVerify, "Skip TLS certificate/hostname verification for Cassandra (INSECURE)")

	fs.StringVar(&cfg.Table, "table", cfg.Table, "Target table name")
	fs.Int64Var(&cfg.TableSize, "table-size", cfg.TableSize, "Number of rows")
	fs.IntVar(&cfg.PayloadSize, "payload-size", cfg.PayloadSize, "Payload size in bytes")

	fs.IntVar(&cfg.Threads, "threads", cfg.Threads, "Number of concurrent workers")
	fs.DurationVar(&cfg.Timeout, "timeout", cfg.Timeout, "Overall command timeout")

	fs.StringVar((*string)(&cfg.Output), "output", string(cfg.Output), "Output format: text|json")
}
