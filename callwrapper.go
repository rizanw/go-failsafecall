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
	// note: TODO
	CacheConfig *InMemCacheConfig
}

type CallWrapper struct {
	callTimeout time.Duration
	sf          *singleflight.Group
}

// New creates CallWrapper
func New(cfg Config) *CallWrapper {
	var (
		callTimeout time.Duration
		sf          *singleflight.Group
	)

	if cfg.CallTimeout > 0 {
		callTimeout = time.Duration(cfg.CallTimeout) * time.Millisecond
	}

	if cfg.Singleflight {
		sf = &singleflight.Group{}
	}

	return &CallWrapper{
		callTimeout: callTimeout,
		sf:          sf,
	}
}

// Call wraps the func call and implement the enabled resiliency patterns
func (cw *CallWrapper) Call(ctx context.Context, key string, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	var (
		res interface{}
		err error
	)

	// context nil prevention and ensure timout call is working
	if ctx == nil {
		ctx = context.Background()
	}

	if cw.sf != nil && key != "" {
		// call singleflight
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

	return res, err
}

func (cw *CallWrapper) call(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	var (
		res    interface{}
		err    error
		cancel context.CancelFunc
	)

	if cw.callTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, cw.callTimeout)
		defer cancel()
	}

	res, err = fn(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}
