package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/rizanw/go-failsafecall"
)

var (
	// use mutex to simulate expensive process during fetching data
	mu sync.Mutex
)

// getPost simulates external call to fetch data from upstream (service/server/db/redis/etc)
func getPost(ctx context.Context, fs *failsafecall.Wrapper, postID int) (map[string]interface{}, error) {
	var (
		url     = fmt.Sprintf("https://jsonplaceholder.typicode.com/posts/%d", postID)
		callKey = fmt.Sprintf("post:%d", postID)
	)

	// use callwrapper to wrap fetching data
	resp, err := fs.Call(ctx, callKey, func(ctx context.Context) (interface{}, error) {
		if postID == 9 {
			// trigger cb becomes half-open and then open
			return nil, errors.New("post not found")
		}

		// simulate fetching data inside this func(ctx context.Context) (interface{}, error):
		var result map[string]interface{}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return nil, err
		}

		// simulating additional wait time
		mu.Lock()
		time.Sleep(2 * time.Second)
		mu.Unlock()

		return result, nil
	}, failsafecall.WithCacheTTL(30))
	if err != nil {
		return nil, err
	}

	// parse response data
	return resp.(map[string]interface{}), err
}
