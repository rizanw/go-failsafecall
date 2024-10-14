package callwrapper

import (
	"errors"
	"time"

	"github.com/sony/gobreaker/v2"
)

type CBConfig struct {
	// OpenTimeoutSec is the period of the open state, after which the state of CircuitBreaker becomes half-open (default: 60).
	OpenTimeoutSec int

	// HalfOpenMaxRequests is the maximum number of requests allowed to pass through when the CircuitBreaker is half-open (default: 1).
	// HalfOpenMaxRequests is also used for calculate minimum success to make CircuitBreaker is being close state again.
	HalfOpenMaxRequests int

	// FailureRatioThreshold is failure threshold percentage before CircuitBreaker is being open state (default: 0.7).
	CloseFailureRatioThreshold float64

	// CloseMinRequestThreshold is the minimum number of request allowed before CircuitBreaker is being open state (default: 10).
	CloseMinRequests int

	// WhitelistedErrors are errors marked as successful process (default: nil).
	// All errors will contribute to make CircuitBreaker is being open state, whitelist the errors will prevent it.
	// example: []error{sql.ErrNoRows} when you need to mark `not found` as successful process.
	WhitelistedErrors []error
}

const (
	defaultCBName = "circuit-breaker"
)

type circuitBreaker struct {
	cb *gobreaker.CircuitBreaker[any]
}

func newCircuitBreaker(cfg *CBConfig) *circuitBreaker {
	var (
		st                         gobreaker.Settings
		OpenTimeout                time.Duration = 60 * time.Second
		HalfOpenMaxRequests        uint32        = 1
		CloseMinRequests           uint32        = 10
		CloseFailureRatioThreshold float64       = 0.7
	)

	st.Name = defaultCBName
	if cfg.OpenTimeoutSec > 0 {
		OpenTimeout = time.Duration(cfg.OpenTimeoutSec) * time.Second
	}
	st.Timeout = OpenTimeout
	if cfg.HalfOpenMaxRequests > 0 {
		HalfOpenMaxRequests = uint32(cfg.HalfOpenMaxRequests)
	}
	st.MaxRequests = HalfOpenMaxRequests
	if cfg.CloseFailureRatioThreshold > 0.0 && cfg.CloseFailureRatioThreshold <= 1.0 {
		CloseFailureRatioThreshold = cfg.CloseFailureRatioThreshold
	}
	if cfg.CloseMinRequests > 0 {
		CloseMinRequests = uint32(cfg.CloseMinRequests)
	}
	st.ReadyToTrip = func(counts gobreaker.Counts) bool {
		failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
		return counts.Requests >= CloseMinRequests && failureRatio >= CloseFailureRatioThreshold
	}
	if cfg.WhitelistedErrors != nil && len(cfg.WhitelistedErrors) > 0 {
		st.IsSuccessful = func(err error) bool {
			for _, whitelistedError := range cfg.WhitelistedErrors {
				if errors.Is(err, whitelistedError) {
					return true
				}
			}
			return false
		}
	}

	return &circuitBreaker{
		cb: gobreaker.NewCircuitBreaker[any](st),
	}
}

func (cb *circuitBreaker) Execute(job func() (interface{}, error)) (interface{}, error) {
	var (
		res interface{}
		err error
	)

	res, err = cb.cb.Execute(job)
	if err != nil {
		return nil, err
	}
	return res, nil
}
