package failsafecall

import (
	"time"

	"github.com/karlseguin/ccache/v3"
)

type InMemCacheConfig struct {
	// MaxSize is the maximum number size to store in the cache (default: 5000)
	MaxSize int

	// Buckets is ccache shards its internal map to provide a greater amount of concurrency.
	// Must be a power of 2 (default: 16).
	Buckets int

	// GetsPerPromote is the number of times an item is fetched before we promote it. For large caches with long TTLs,
	// it normally isn't necessary to promote an item after every fetch (default: 3)
	GetsPerPromote int

	// ItemsToPrune is the number of items to prune when we hit MaxSize.
	// Freeing up more than 1 slot at a time improved performance (default: 500)
	ItemsToPrune int

	// TTLSec is the number of duration that a cached item is considered valid or fresh. (default: 1)
	TTLSec int
}

type cache struct {
	cache      *ccache.Cache[any]
	defaultTTL time.Duration
}

func newCache(cfg *InMemCacheConfig) *cache {
	var (
		maxSize        int64  = 5000
		buckets        uint32 = 16
		getsPerPromote int32  = 3
		itemsToPrune   uint32 = 500
		ttlSec         int    = 1
	)

	if cfg.MaxSize > 0 {
		maxSize = int64(cfg.MaxSize)
	}
	if cfg.Buckets > 0 {
		buckets = uint32(cfg.Buckets)
	}
	if cfg.GetsPerPromote > 0 {
		getsPerPromote = int32(cfg.GetsPerPromote)
	}
	if cfg.ItemsToPrune > 0 {
		itemsToPrune = uint32(cfg.ItemsToPrune)
	}
	if cfg.TTLSec > 0 {
		ttlSec = cfg.TTLSec
	}

	return &cache{
		cache: ccache.New[any](
			ccache.Configure[any]().
				MaxSize(maxSize).
				Buckets(buckets).
				GetsPerPromote(getsPerPromote).
				ItemsToPrune(itemsToPrune),
		),
		defaultTTL: time.Duration(ttlSec) * time.Second,
	}
}

func (c *cache) Set(key string, val interface{}, ttl time.Duration) {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	c.cache.Set(key, val, ttl)
}

func (c *cache) Get(key string) (res interface{}, isExist bool) {
	item := c.cache.Get(key)
	if item == nil || item.Expired() {
		return nil, false
	}

	return item.Value(), true
}
