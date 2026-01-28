package workload

import (
	"time"

	"github.com/spf13/pflag"

	"tidb-benchmarks/pkg/config"
)

type Kind string

const (
	KindReadOnly  Kind = "read-only"
	KindWriteOnly Kind = "write-only"
	KindMixed     Kind = "mixed"
)

func BindRunFlags(fs *pflag.FlagSet, cfg *config.Config) {
	fs.DurationVar(&cfg.Time, "time", cfg.Time, "Workload duration (e.g. 30s)")
}

func BindMixedFlags(fs *pflag.FlagSet, cfg *config.Config) {
	fs.Float64Var(&cfg.ReadRatio, "read-ratio", cfg.ReadRatio, "Read ratio for mixed workload (0..1)")
	fs.DurationVar(&cfg.Warmup, "warmup", cfg.Warmup, "Warmup duration before measuring")
}

func clampRatio(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

func effectiveWarmup(d time.Duration) time.Duration {
	if d < 0 {
		return 0
	}
	return d
}
