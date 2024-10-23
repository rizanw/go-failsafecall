package failsafecall

import "time"

type callOptions[T any] struct {
	TimeoutDeadline time.Duration
	CacheTTL        time.Duration
}

type CallOption func(option *callOptions[any])

// WithTimeoutDeadline func overrides the TimeoutDeadline configuration
func WithTimeoutDeadline[T int | int64](timeoutMs T) CallOption {
	return func(o *callOptions[any]) {
		o.TimeoutDeadline = time.Duration(timeoutMs) * time.Millisecond
	}
}

// WithCacheTTL func overrides the cache TTL configuration
func WithCacheTTL[T int | int32 | int64](TTLSec T) CallOption {
	return func(o *callOptions[any]) {
		o.CacheTTL = time.Duration(TTLSec) * time.Second
	}
}
