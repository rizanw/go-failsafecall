package main

import (
	"context"
	"fmt"
	"time"

	"github.com/rizanw/go-callwrapper"
)

// simulateCircuitBreaker is an example when you expect to have circuit breaker.
func simulateCircuitBreaker() {
	var (
		ctx     = context.Background()
		now     = time.Now()
		postIDs = []int{9, 9, 3, 1, 9, 3, 9, 1, 9, 9}
	)

	cw := callwrapper.New(callwrapper.Config{
		CBConfig: &callwrapper.CBConfig{
			OpenTimeoutSec:             10,
			HalfOpenMaxRequests:        2,
			CloseFailureRatioThreshold: 0.5,
			CloseMinRequests:           5,
		},
	})

	// simulate concurrent request happens
	for _, postID := range postIDs {
		// fetching data from repository
		data, err := getPost(ctx, cw, postID)
		if err != nil {
			fmt.Printf(">PostID:%v, Error:%v\n", postID, err)
		} else {
			fmt.Printf(">PostID:%v, Title:%v\n", postID, data["title"])
		}
	}

	fmt.Println("[simulateTimeoutCall] Time taken:", time.Since(now))
}
