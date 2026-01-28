package metrics

import (
	"encoding/json"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
)

type Summary struct {
	Name string `json:"name"`

	Start time.Time     `json:"start"`
	End   time.Time     `json:"end"`
	Dur   time.Duration `json:"duration"`

	Ops    int64 `json:"ops"`
	Errors int64 `json:"errors"`
	Bytes  int64 `json:"bytes"`

	AvgMs  float64 `json:"avg_ms"`
	P50Ms  float64 `json:"p50_ms"`
	P95Ms  float64 `json:"p95_ms"`
	P99Ms  float64 `json:"p99_ms"`
	P999Ms float64 `json:"p999_ms"`

	QPS float64 `json:"qps"`
	BPS float64 `json:"bytes_per_sec"`
}

type Recorder struct {
	h *hdrhistogram.Histogram

	start time.Time
	end   time.Time

	ops    int64
	errors int64
	bytes  int64
}

func NewRecorder() *Recorder {
	// 1us..60s, 3 significant figures.
	return &Recorder{h: hdrhistogram.New(1, int64((60 * time.Second).Microseconds()), 3)}
}

func (r *Recorder) Start(t time.Time) { r.start = t }
func (r *Recorder) End(t time.Time)   { r.end = t }

func (r *Recorder) Record(d time.Duration, nbytes int, ok bool) {
	us := d.Microseconds()
	if us < 1 {
		us = 1
	}
	_ = r.h.RecordValue(us)
	r.ops++
	if !ok {
		r.errors++
	}
	r.bytes += int64(nbytes)
}

func (r *Recorder) Merge(other *Recorder) {
	r.h.Merge(other.h)
	r.ops += other.ops
	r.errors += other.errors
	r.bytes += other.bytes
	if r.start.IsZero() || (!other.start.IsZero() && other.start.Before(r.start)) {
		r.start = other.start
	}
	if r.end.IsZero() || other.end.After(r.end) {
		r.end = other.end
	}
}

func (r *Recorder) Summary(name string) Summary {
	start := r.start
	end := r.end
	if start.IsZero() {
		start = time.Now()
	}
	if end.IsZero() || end.Before(start) {
		end = start
	}
	dur := end.Sub(start)
	if dur <= 0 {
		dur = time.Nanosecond
	}

	avgUs := float64(r.h.Mean())
	q := func(p float64) float64 {
		return float64(r.h.ValueAtQuantile(p)) / 1000.0
	}

	qps := float64(r.ops) / dur.Seconds()
	bps := float64(r.bytes) / dur.Seconds()

	return Summary{
		Name:   name,
		Start:  start,
		End:    end,
		Dur:    dur,
		Ops:    r.ops,
		Errors: r.errors,
		Bytes:  r.bytes,
		AvgMs:  avgUs / 1000.0,
		P50Ms:  q(50),
		P95Ms:  q(95),
		P99Ms:  q(99),
		P999Ms: q(99.9),
		QPS:    qps,
		BPS:    bps,
	}
}

func (s Summary) MarshalJSON() ([]byte, error) {
	type Alias Summary
	return json.Marshal(&struct {
		Alias
		DurationMs float64 `json:"duration_ms"`
	}{
		Alias:      Alias(s),
		DurationMs: float64(s.Dur) / float64(time.Millisecond),
	})
}
