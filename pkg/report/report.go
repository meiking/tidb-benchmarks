package report

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"tidb-benchmarks/pkg/config"
	"tidb-benchmarks/pkg/metrics"
)

func Print(format config.OutputFormat, s metrics.Summary) error {
	switch format {
	case config.OutputJSON:
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(s)
	case config.OutputText, "":
		fmt.Printf("Name: %s\n", s.Name)
		fmt.Printf("Duration: %s\n", s.Dur.Round(time.Millisecond))
		fmt.Printf("Ops: %d\n", s.Ops)
		fmt.Printf("Errors: %d\n", s.Errors)
		fmt.Printf("Bytes: %d\n", s.Bytes)
		fmt.Printf("QPS: %.2f\n", s.QPS)
		fmt.Printf("BPS: %.2f\n", s.BPS)
		fmt.Printf("Latency(ms): avg=%.3f p50=%.3f p95=%.3f p99=%.3f p999=%.3f\n", s.AvgMs, s.P50Ms, s.P95Ms, s.P99Ms, s.P999Ms)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}
