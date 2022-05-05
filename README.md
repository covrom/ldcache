# ldcache
Loading cache with Go generics 1.18+

Example of usage:

```go
var idCache *ldcache.Cache[uuid.UUID, *User]

idCache = ldcache.NewCache(
    tcache.CacheParams[uuid.UUID, *User]{
        Loader: func(ctx context.Context, key uuid.UUID) (*User, error) {
                return UserDataFromStoreByID(key)
            },
        Size:    cacheCount,
        Expires: time.Hour * 12,
    },
)

vuser, err := idCache.Get(ctx, id)
```