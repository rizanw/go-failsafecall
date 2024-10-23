package main

import (
	"context"
	"fmt"
	"time"

	"github.com/rizanw/go-failsafecall"
)

// simulateInMemoryCache is an example when you expect to reduce/compress identical request to upstream
func simulateInMemoryCache() {
	var (
		ctx     = context.Background()
		now     = time.Now()
		postIDs = []int{1, 1, 3, 2, 1, 3, 1, 2, 3, 1}
	)

	fs := failsafecall.New(failsafecall.Config{
		CacheConfig: &failsafecall.InMemCacheConfig{TTLSec: 60}, // enable in-memory cache feature with default configuration except ttl
	})

	// simulate burst request happens:
	// by enabling in-memory cache inside failsafecall wrapper, the request into external will be blocked
	// when the previous identical data exist inside cache
	for _, postID := range postIDs {
		// fetching data from repository
		data, err := getPost(ctx, fs, postID)
		if err != nil {
			fmt.Printf(">PostID:%v, Error:%v\n", postID, err)
		} else {
			fmt.Printf(">PostID:%v, Title:%v\n", postID, data["title"])
		}
	}

	fmt.Println("[simulateTimeoutCall] Time taken:", time.Since(now))
}
