package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rizanw/go-failsafecall"
)

// simulateSingleflight is an example when you expect to reduce/compress identical request to upstream
func simulateSingleflight() {
	var (
		ctx     = context.Background()
		now     = time.Now()
		postIDs = []int{1, 2, 3, 1, 2, 3, 1, 1, 2, 4}
	)

	fsc := failsafecall.New(failsafecall.Config{
		Singleflight: true, // enable singleflight feature
	})

	// simulate concurrent request happens
	var wg sync.WaitGroup
	for _, postID := range postIDs {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			// fetching data from repository
			data, err := getPost(ctx, fsc, id)
			if err != nil {
				fmt.Printf(">PostID:%v, Error:%v\n", id, err)
			} else {
				fmt.Printf(">PostID:%v, Title:%v\n", id, data["title"])
			}
		}(postID)
	}

	// since inside repo we expect each request will respond about 2s
	// when singleflight: false (disabled feature) will return the response one by one
	// example:
	// - in: 1, 2, 3, 1, 2, 3, 1, 1, 2, 4
	// - out: 1, 2, 3, 1, 2, 3, 1, 1, 2, 4 (potentially random)
	// when singleflight: true (enabled feature) will return in a group with the same key
	// example:
	// - in:	1, 2, 3, 1, 2, 3, 1, 1, 2, 4
	// - out:	3, 3, 2, 2, 2, 4, 1, 1, 1, 1 (potentially random, but grouped)
	wg.Wait()
	fmt.Println("[simulateTimeoutCall] Time taken:", time.Since(now))
}
