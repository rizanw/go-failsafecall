package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rizanw/go-callwrapper"
)

// simulateTimeoutCall is an example when you expect to have request timeout for external service
func simulateTimeoutCall() {
	var (
		ctx     = context.Background()
		now     = time.Now()
		postIDs = []int{1, 2, 3, 1, 2, 3, 1, 1, 2, 4}
	)

	cw := callwrapper.New(callwrapper.Config{
		CallTimeout: 500, // 500ms
	})

	// simulate concurrent request happens
	var wg sync.WaitGroup
	for _, postID := range postIDs {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			// fetching data from repository
			data, err := getPost(ctx, cw, id)
			if err != nil {
				// since inside repo we expect each request will respond about 2s and our timeout is 5ms
				// all fetching data will return `context deadline exceeded`
				fmt.Printf(">PostID:%v, Error:%v\n", id, err)
			} else {
				fmt.Printf(">PostID:%v, Title:%v\n", id, data["title"])
			}
		}(postID)
	}

	wg.Wait()
	fmt.Println("[simulateTimeoutCall] Time taken:", time.Since(now))
}
