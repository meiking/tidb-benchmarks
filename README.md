# mysql-cassandra-bench (Go)

A small sysbench-style benchmarking tool to compare MySQL vs Cassandra for:

- `prepare` (data preparation)
- `read-only`
- `write-only`
- `mixed`

It reports latency distribution (avg/p95/p99/p999), throughput, and total processed data.

## Build

```bash
go mod tidy
go build -o bench ./cmd/bench
```

## Quick start

### MySQL

```bash
./bench prepare \
  --db mysql \
  --mysql-dsn 'root:password@tcp(127.0.0.1:3306)/test?parseTime=true&multiStatements=true' \
  --mysql-tls-skip-verify=false \
  --table sbtest \
  --table-size 100000 \
  --threads 16

./bench run read-only \
  --db mysql \
  --mysql-dsn 'root:password@tcp(127.0.0.1:3306)/test?parseTime=true&multiStatements=true' \
  --mysql-tls-skip-verify=false \
  --table sbtest \
  --table-size 100000 \
  --threads 64 \
  --time 30s
```

### Cassandra

```bash
./bench prepare \
  --db cassandra \
  --cassandra-hosts '127.0.0.1' \
  --cassandra-keyspace 'bench' \
  --cassandra-tls-skip-verify=false \
  --table sbtest \
  --table-size 100000 \
  --threads 16

./bench run mixed \
  --db cassandra \
  --cassandra-hosts '127.0.0.1' \
  --cassandra-keyspace 'bench' \
  --cassandra-tls-skip-verify=false \
  --table sbtest \
  --table-size 100000 \
  --threads 64 \
  --time 30s \
  --read-ratio 0.5
```

## Output

- Default is human-readable text.
- Use `--output json` to emit a single JSON object (useful for CI).

## Notes

- For fairness, try to keep schema, payload size, and consistency settings comparable.
- Cassandra consistency defaults to `LOCAL_QUORUM` (configurable).
