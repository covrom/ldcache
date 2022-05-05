// Package cache provides implementations of loading cache,
// including support for LRU and LFU.
package ldcache

import "context"

// Key is any value which is comparable.
type Key comparable

// Value is any value.
type Value any

// SomeCache is a key-value cache which entries are added and stayed in the
// cache until either are evicted or manually invalidated.
type SomeCache[K Key, V Value] interface {
	// GetIfPresent returns value associated with Key or (nil, false)
	// if there is no cached value for Key.
	GetIfPresent(context.Context, K) (V, bool)

	// Put associates value with Key. If a value is already associated
	// with Key, the old one will be replaced with Value.
	Put(context.Context, K, V)

	// Invalidate discards cached value of the given Key.
	Invalidate(context.Context, K)

	// InvalidateAll discards all entries.
	InvalidateAll(context.Context)

	// Close implements io.Closer for cleaning up all resources.
	// Users must ensure the cache is not being used before closing or
	// after closed.
	Close() error
}

// Func is a generic callback for entry events in the cache.
type Func[K Key, V Value] func(K, V)

// LoadingCache is a cache with values are loaded automatically and stored
// in the cache until either evicted or manually invalidated.
type LoadingCache[K Key, V Value] interface {
	SomeCache[K, V]

	// Get returns value associated with Key or call underlying LoaderFunc
	// to load value if it is not present.
	Get(context.Context, K) (V, error)
	// SetLoader changes the underlying loader func
	SetLoader(LoaderFunc[K, V])
	// Refresh loads new value for Key. If the Key already existed, the previous value
	// will continue to be returned by Get while the new value is loading.
	// If Key does not exist, this function will block until the value is loaded.
	Refresh(context.Context, K) error
}

// LoaderFunc retrieves the value corresponding to given Key.
// You must ensure the context is not canceled before use it in database queries!
type LoaderFunc[K Key, V Value] func(context.Context, K) (V, error)
