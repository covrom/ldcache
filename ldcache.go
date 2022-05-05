package ldcache

import (
	"context"
	"errors"
	"sync"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/Code-Hex/go-generics-cache/policy/lfu"
	"github.com/Code-Hex/go-generics-cache/policy/lru"
)

type CacheType int

const (
	LRU CacheType = iota
	LFU
)

var _ LoadingCache[string, string] = &Cache[string, string]{}

type Cache[K Key, V Value] struct {
	c   *cache.Cache[K, V]
	exp time.Duration
	mu  sync.RWMutex
	lf  LoaderFunc[K, V]
}

type CacheParams[K Key, V Value] struct {
	Type    CacheType     // default LRU
	Size    int           // default 10000
	Expires time.Duration // default never
	Loader  LoaderFunc[K, V]
}

func NewCache[K Key, V Value](params CacheParams[K, V]) *Cache[K, V] {
	if params.Size == 0 {
		params.Size = 10000
	}
	ret := &Cache[K, V]{
		exp: params.Expires,
		lf:  params.Loader,
	}
	if params.Type == LRU {
		ret.c = cache.New(
			cache.AsLRU[K, V](
				lru.WithCapacity(params.Size),
			),
		)
	}
	if params.Type == LFU {
		ret.c = cache.New(
			cache.AsLFU[K, V](
				lfu.WithCapacity(params.Size),
			),
		)
	}
	return ret
}

func (tc *Cache[K, V]) GetIfPresent(ctx context.Context, key K) (V, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	gota, aok := tc.c.Get(key)
	return gota, aok
}

func (tc *Cache[K, V]) Put(ctx context.Context, key K, val V) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.c.Set(key, val, cache.WithExpiration(tc.exp))
}

func (tc *Cache[K, V]) Invalidate(ctx context.Context, key K) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.c.Delete(key)
}

func (tc *Cache[K, V]) InvalidateAll(ctx context.Context) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	ks := tc.c.Keys()
	for _, k := range ks {
		tc.c.Delete(k)
	}
}

func (tc *Cache[K, V]) Close() error {
	return nil
}

var ErrNotFound = errors.New("key not found in cache")

func (tc *Cache[K, V]) Get(ctx context.Context, key K) (V, error) {
	tc.mu.RLock()
	v, ok := tc.c.Get(key)
	tc.mu.RUnlock()
	if ok {
		return v, nil
	}
	tc.mu.Lock()
	defer tc.mu.Unlock()
	v, ok = tc.c.Get(key)
	if ok {
		return v, nil
	} else if tc.lf == nil {
		return v, ErrNotFound
	}
	v, err := tc.lf(ctx, key)
	if err != nil {
		return v, err
	}

	tc.c.Set(key, v, cache.WithExpiration(tc.exp))
	return v, nil
}

func (tc *Cache[K, V]) SetLoader(lf LoaderFunc[K, V]) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.lf = lf
}

func (tc *Cache[K, V]) Refresh(ctx context.Context, key K) error {
	if tc.lf == nil {
		return nil
	}

	tc.mu.Lock()
	defer tc.mu.Unlock()

	v, err := tc.lf(ctx, key)
	if err != nil {
		return err
	}

	tc.c.Set(key, v, cache.WithExpiration(tc.exp))
	return nil
}
