package callwrapper

import (
	"context"
	"time"

	"golang.org/x/sync/singleflight"
)

// Config for CallWrapper configuration
type Config struct {
	// CallTimeout set context timeout in milliseconds
	// note: use client with context to make it works!
	CallTimeout int64

	// Singleflight option
	// note: singleflight won't work without key, ensure provide unique key when using the `Call` function
	Singleflight bool

	// Circuit Breaker configuration
	// note: TODO
	CBConfig *CBConfig

	// In-Memory Cache configuration
	// note:
	// - setup config with nil means disable feature, empty config means using default configuration
	// - to specify TTL on each cw.Call use WithCacheTTL func
	CacheConfig *InMemCacheConfig
}

type CallWrapper struct {
	callTimeout time.Duration
	sf          *singleflight.Group
	cache       *cache
}

// New creates CallWrapper
func New(cfg Config) *CallWrapper {
	var (
		callTimeout time.Duration
		sf          *singleflight.Group
		c           *cache
	)

	if cfg.CallTimeout > 0 {
		callTimeout = time.Duration(cfg.CallTimeout) * time.Millisecond
	}

	if cfg.Singleflight {
		sf = &singleflight.Group{}
	}

	if cfg.CacheConfig != nil {
		c = newCache(cfg.CacheConfig)
	}

	return &CallWrapper{
		callTimeout: callTimeout,
		sf:          sf,
		cache:       c,
	}
}

// Call wraps the func call and implement the enabled resiliency patterns
func (cw *CallWrapper) Call(ctx context.Context, key string, fn func(ctx context.Context) (interface{}, error), opts ...CallOption) (interface{}, error) {
	var (
		res interface{}
		err error
		co  = &callOptions[any]{}
	)

	// context nil prevention and ensure timout call is working
	if ctx == nil {
		ctx = context.Background()
	}

	// apply CallOption function into callOptions instance
	for _, opt := range opts {
		opt(co)
	}

	// set context timeout
	if cw.callTimeout > 0 || co.TimeoutDeadline > 0 {
		var (
			timeout time.Duration = cw.callTimeout
			cancel  context.CancelFunc
		)

		if co.TimeoutDeadline > 0 {
			timeout = co.TimeoutDeadline
		}

		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// return data from cache if exist
	if cw.cache != nil && key != "" {
		if res, exist := cw.cache.Get(key); exist {
			return res, nil
		}
	}

	// do the call with or without singleflight
	if cw.sf != nil && key != "" {
		// call with singleflight
		res, err = cw.call(ctx, func(ctx context.Context) (interface{}, error) {
			sfRes, sfErr, _ := cw.sf.Do(key, func() (interface{}, error) {
				return fn(ctx)
			})
			return sfRes, sfErr
		})
	} else {
		// call without singleflight
		res, err = cw.call(ctx, func(ctx context.Context) (interface{}, error) {
			return fn(ctx)
		})
	}
	if err != nil {
		return nil, err
	}

	// set cache on success call
	if cw.cache != nil && key != "" {
		cw.cache.Set(key, res, co.CacheTTL)
	}

	return res, err
}

func (cw *CallWrapper) call(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	var (
		res interface{}
		err error
	)

	res, err = fn(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}
