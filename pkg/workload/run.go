package workload

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"tidb-benchmarks/pkg/config"
	"tidb-benchmarks/pkg/db"
	"tidb-benchmarks/pkg/metrics"
	"tidb-benchmarks/pkg/util"
)

func Run(ctx context.Context, client db.Client, cfg config.Config, kind Kind) (metrics.Summary, error) {
	if cfg.TableSize <= 0 {
		return metrics.Summary{}, fmt.Errorf("table-size must be > 0")
	}
	if cfg.Threads <= 0 {
		return metrics.Summary{}, fmt.Errorf("threads must be > 0")
	}
	if cfg.Time <= 0 {
		return metrics.Summary{}, fmt.Errorf("time must be > 0")
	}

	if err := client.PrepareSchema(ctx, cfg); err != nil {
		return metrics.Summary{}, err
	}

	readRatio := clampRatio(cfg.ReadRatio)
	warmup := effectiveWarmup(cfg.Warmup)

	payload := util.MakePayload(cfg.PayloadSize)

	endWarmup := time.Now().Add(warmup)
	startMeasure := endWarmup
	endMeasure := startMeasure.Add(cfg.Time)

	var mu sync.Mutex
	global := metrics.NewRecorder()

	eg, egctx := errgroup.WithContext(ctx)
	eg.SetLimit(cfg.Threads)

	for i := 0; i < cfg.Threads; i++ {
		workerID := i
		eg.Go(func() error {
			rng := util.NewSplitMix64(uint64(time.Now().UnixNano()) + uint64(workerID)*104729)
			local := metrics.NewRecorder()
			measuring := false

			for {
				now := time.Now()
				if now.After(endMeasure) {
					break
				}
				if !measuring && now.After(endWarmup) {
					measuring = true
					local.Start(startMeasure)
				}

				id := 1 + rng.Int63n(cfg.TableSize)
				k := rng.Int63n(cfg.TableSize)

				doRead := false
				switch kind {
				case KindReadOnly:
					doRead = true
				case KindWriteOnly:
					doRead = false
				case KindMixed:
					doRead = float64(rng.Next()%10000)/10000.0 < readRatio
				default:
					return fmt.Errorf("unsupported workload kind: %s", kind)
				}

				if doRead {
					t0 := time.Now()
					payloadOut, err := client.Read(egctx, cfg, id)
					if measuring {
						local.Record(time.Since(t0), len(payloadOut), err == nil)
					}
					if err != nil {
						return err
					}
					continue
				}

				t0 := time.Now()
				err := client.Update(egctx, cfg, id, k, payload)
				if measuring {
					local.Record(time.Since(t0), len(payload), err == nil)
				}
				if err != nil {
					return err
				}
			}

			if measuring {
				local.End(endMeasure)
				mu.Lock()
				global.Merge(local)
				mu.Unlock()
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		global.End(time.Now())
		return global.Summary(fmt.Sprintf("%s/%s", kind, client.Name())), err
	}

	if global.Summary("tmp").Ops == 0 {
		// Warmup may exceed the total runtime.
		global.Start(startMeasure)
		global.End(endMeasure)
	}
	return global.Summary(fmt.Sprintf("%s/%s", kind, client.Name())), nil
}
