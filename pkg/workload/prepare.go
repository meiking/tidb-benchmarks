package workload

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"tidb-benchmarks/pkg/config"
	"tidb-benchmarks/pkg/db"
	"tidb-benchmarks/pkg/metrics"
	"tidb-benchmarks/pkg/util"
)

func Prepare(ctx context.Context, client db.Client, cfg config.Config) (metrics.Summary, error) {
	if cfg.TableSize <= 0 {
		return metrics.Summary{}, fmt.Errorf("table-size must be > 0")
	}
	if cfg.Threads <= 0 {
		return metrics.Summary{}, fmt.Errorf("threads must be > 0")
	}

	if err := client.PrepareSchema(ctx, cfg); err != nil {
		return metrics.Summary{}, err
	}
	if err := client.Truncate(ctx, cfg); err != nil {
		return metrics.Summary{}, err
	}

	payload := util.MakePayload(cfg.PayloadSize)

	var mu sync.Mutex
	global := metrics.NewRecorder()
	start := time.Now()
	global.Start(start)

	var nextID int64
	eg, egctx := errgroup.WithContext(ctx)
	eg.SetLimit(cfg.Threads)

	for i := 0; i < cfg.Threads; i++ {
		workerID := i
		eg.Go(func() error {
			rng := util.NewSplitMix64(uint64(time.Now().UnixNano()) + uint64(workerID)*7919)
			local := metrics.NewRecorder()
			local.Start(start)
			for {
				id := atomic.AddInt64(&nextID, 1)
				if id > cfg.TableSize {
					local.End(time.Now())
					mu.Lock()
					global.Merge(local)
					mu.Unlock()
					return nil
				}
				k := rng.Int63n(cfg.TableSize)
				t0 := time.Now()
				err := client.Insert(egctx, cfg, id, k, payload)
				local.Record(time.Since(t0), len(payload), err == nil)
				if err != nil {
					return err
				}
			}
		})
	}

	if err := eg.Wait(); err != nil {
		global.End(time.Now())
		return global.Summary(fmt.Sprintf("prepare/%s", client.Name())), err
	}
	global.End(time.Now())
	return global.Summary(fmt.Sprintf("prepare/%s", client.Name())), nil
}
