package main

import "fmt"

func main() {
	// example for using timeout
	simulateTimeoutCall()
	fmt.Println("=====================")

	// example for using singleflight
	simulateSingleflight()
	fmt.Println("=====================")

	// example for using in memory cache
	simulateInMemoryCache()
	fmt.Println("=====================")

	// example for using circuit breaker
	simulateCircuitBreaker()
}
