# go-failsafecall

go-failsafecall is basically a wrapper function designed to perform external calls and implementing commonly used
distributed system pattern best practices to achieve stability and resilience:

## Features

- [Circuit Breaker Pattern](#circuit-breaker-pattern)
- [LRU (Least Recently Used) In-Memory Cache Pattern](#lru-in-memory-cache-pattern)
- [Singleflight Pattern](#singleflight-pattern)
- [Request Timeout](#request-timeout-deadline)

## Circuit Breaker Pattern

The Circuit Breaker pattern is a design pattern used in software engineering to improve the stability and resilience of
applications, particularly in distributed systems. It acts like an electrical circuit breaker, preventing an application
from trying to execute an operation that is likely to fail.

### When to Use this Pattern

**Use this pattern:**

- To prevent an application from attempting to invoke an external service or access a shared resource if this operation
  is highly likely to fail.

**This pattern might not be suitable:**

- For handling access to local private resources in an application, such as in-memory data structure. In this
  environment, using a circuit breaker would simply add overhead to your system.
- As a substitute for handling exceptions in the business logic of your applications.

<p align="right">(
<a href="https://learn.microsoft.com/en-us/previous-versions/msp-n-p/dn589784(v=pandp.10)">reference</a>
)</p>

## LRU In-Memory Cache Pattern

In-memory caching patterns are techniques used to temporarily store frequently accessed data in memory to improve
application performance and reduce latency. In this failsafecall package will implement Least Recently Used (LRU)
policy.
The LRU Cache operates on the principle that the data most recently accessed is likely to be accessed again in the near
future. By evicting the least recently accessed items first, LRU Cache ensures that the most relevant data remains
available in the cache.

### When to Use this Pattern

**Use this pattern:**

- When certain data is accessed repeatedly within a short period.
- If retrieving data involves costly computations or time-consuming operations, caching can help reduce these costs.
- Data that doesn't change often or is predictable (like configurations) is ideal for caching.
- In scenarios where latency is critical, such as real-time applications, caching can lead to noticeable improvements.

**This pattern might not be suitable:**

- On highly dynamic data or when the source data changes frequently, caching can lead to stale data being served to
  users unless properly invalidated. Or in write-heavy applications, caching might not provide significant benefits, as
  data is often being changed rather than read.

<p align="right">(
<a href="https://redis.io/glossary/lru-cache/">reference</a>
)</p>

## Singleflight Pattern

Singleflight pattern is a concurrency pattern designed to prevent duplicate function calls for the same key when
multiple goroutines request the same resource. It ensures that the function is executed only once, and the result is
shared among all callers.

### When to Use this Pattern

**Use this pattern:**

- To reducing load, it can be used to reduce load on external services or databases by ensuring that requests for the
  same data are consolidated.
- To preventing duplicate work, use singleflight when you have expensive computations or any function
  that should only be executed once for a given key, even if requested by multiple goroutines concurrently.

**This pattern might not be suitable:**

- If the requests are unique and have different parameters or need unique handling, single-flight might not be
  appropriate since it groups requests. It also adds complexity so if the benefits call not outweigh the overhead.

<p align="right">(
<a href="https://victoriametrics.com/blog/go-singleflight/">reference-1</a>
<a href="https://www.codingexplorations.com/blog/understanding-singleflight-in-golang-a-solution-for-eliminating-redundant-work">reference-2</a>
)</p>

## Request Timeout Deadline

It is not a pattern, but to manage deadlines, cancellation signals, and request-scoped values across API boundaries.
Using golang context with a timeout is a context that automatically cancels after a specified duration. This is useful
for operations that may take an uncertain amount of time and helps prevent resource leaks and unresponsive programs.

<p align="right">(
<a href="https://pkg.go.dev/context#WithTimeout">reference</a>
)</p>

## Architectural Flow Diagram Design

when all features are enabled, the execution func described as follows:

```mermaid
sequenceDiagram
    participant fs as fs.Call
    participant td as TimeoutDeadline
    participant c as In-Mem Cache
    participant sf as Singleflight
    participant cb as CircuitBreaker
    participant fn as fn function
    fs ->> td: Execute
    td ->> td: Set Context Deadline
    td ->> c: Execute
    c -->> fs: Return data if cache exist
    c ->> sf: Execute
    sf ->> cb: Execute
    cb -->> sf: Return Error if Breaker open
    cb ->> fn: Execute
    fn ->> cb: Result or Error
    cb ->> sf: Result or Error
    sf ->> c: Result or Error
    c ->> c: Set Cache if Result
    c ->> fs: Result or Error
```

## Quick Start

the failsafecall provides a simple way with 2 steps:

1. create the wrapper instance as `fs`

```go

fs := failsafecall.New(failsafecall.Config{
    TimeoutDeadline: 500, // in milliseconds
    Singleflight: true,
    CBConfig: &failsafecall.CBConfig{
        OpenTimeoutSec: 60, // in seconds 
        HalfOpenMaxRequests: 2,
        CloseFailureRatioThreshold: 0.5,
        CloseMinRequests: 10,
        WhitelistedErrors: []error{sql.ErrNoRows},
    },
    InMemCacheConfig: &failsafecall.InMemCacheConfig{
        TTLSec: 3600, // 1hour in seconds
    }
})

```

2. use the `fs` instance to perform external call

```go
resp, err := fs.Call(ctx, callKey, func (ctx context.Context) (interface{}, error) {
    return getData(ctx)
})
```

## Configuration

below is list of available configuration:

### Wrapper Configuration

| Key          | type              | Description                           |
|--------------|-------------------|---------------------------------------|
| CallTimeout  | int               | set context timeout in milliseconds   |
| Singleflight | bool              | toggle to enable singleflight feature |
| CBConfig     | *CBConfig         | Circuit Breaker configuration         |
| CacheConfig  | *InMemCacheConfig | In-Memory Cache configuration         |

### Circuit Breaker Configuration

| Key                        | type    | Description                                                                                              |
|----------------------------|---------|----------------------------------------------------------------------------------------------------------|
| OpenTimeoutSec             | int     | the period of the open state, after which the state of CircuitBreaker becomes half-open (default: 60)    |
| HalfOpenMaxRequests        | int     | the maximum number of requests allowed to pass through when the CircuitBreaker is half-open (default: 1) |
| CloseFailureRatioThreshold | float64 | failure threshold percentage before CircuitBreaker is being open state (default: 0.5)                    |
| CloseMinRequests           | int     | the minimum number of request allowed before CircuitBreaker is being open state (default: 10)            |
| WhitelistedErrors          | []error | errors marked as successful process (default: nil)                                                       |

### In-Memory Cache Configuration

| Key            | type | Description                                                                                                                                                                 |
|----------------|------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| MaxSize        | int  | the maximum number size to store in the cache (default: 5000)                                                                                                               |
| Buckets        | int  | ccache shards its internal map to provide a greater amount of concurrency. Must be a power of 2 (default: 16).                                                              |
| GetsPerPromote | int  | the number of times an item is fetched before we promote it. For large caches with long TTLs, it normally isn't necessary to promote an item after every fetch (default: 3) |
| ItemsToPrune   | int  | ItemsToPrune is the number of items to prune when we hit MaxSize. Freeing up more than 1 slot at a time improved performance (default: 500)                                 |
| TTLSec         | int  | the number of duration that a cached item is considered valid or fresh. (default: 1)                                                                                        |

## Option

Option use for update or override initiated configuration for specific use-cases

### WithTimeoutDeadline(timeoutMs)

to set CallTimeout context in milliseconds for specific call only.

```go
resp, err := fs.Call(ctx, key, GetData, failsafecall.WithTimeoutDeadline(30))
if err != nil {
    return nil, err
}
```

### WithCacheTTL(TTLSec)

to set TTL in-memory cache in seconds for specific call only.

```go
resp, err := fs.Call(ctx, key, GetData, failsafecall.WithCacheTTL(30))
if err != nil {
    return nil, err
}
```

## Example Usage

check example folder to see detailed implementation use cases.

- how to perform external call with
  failsafecall ([example](https://github.com/rizanw/go-failsafecall/blob/main/example/repo.go))
- when you need to have set deadline
  call ([example usage](https://github.com/rizanw/go-failsafecall/blob/main/example/ttl.go))
- when you need to reduce upstream load with
  singleflight ([example usage](https://github.com/rizanw/go-failsafecall/blob/main/example/singleflight.go))
- when you need to fetch frequent access and rarely changes data using in-memory
  cache ([example usage](https://github.com/rizanw/go-failsafecall/blob/main/example/cache.go))
- when you need to prevent likely fail request with
  circuit-breaker ([example usage](https://github.com/rizanw/go-failsafecall/blob/main/example/circuitbreaker.go))
