package util

import "time"

// SplitMix64 is a small, fast PRNG suitable for benchmarking workloads.
// It is not intended for cryptographic use.
type SplitMix64 struct {
	x uint64
}

func NewSplitMix64(seed uint64) *SplitMix64 {
	if seed == 0 {
		seed = uint64(time.Now().UnixNano())
	}
	return &SplitMix64{x: seed}
}

func (r *SplitMix64) Next() uint64 {
	r.x += 0x9e3779b97f4a7c15
	z := r.x
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

func (r *SplitMix64) Int63n(n int64) int64 {
	if n <= 0 {
		return 0
	}
	return int64(r.Next() % uint64(n))
}
