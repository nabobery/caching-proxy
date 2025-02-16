package cache

import (
	"log"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

var InMemoryCache *gocache.Cache

func init() {
	// Initialize the in-memory cache with a default expiration of 5 minutes and a cleanup interval of 10 minutes.
	InMemoryCache = gocache.New(5*time.Minute, 10*time.Minute)
}

// ClearCache flushes all the items from the cache.
func ClearCache() {
	if InMemoryCache == nil {
		log.Println("Cache not initialized.")
		return
	}
	InMemoryCache.Flush()
	log.Println("In-memory cache cleared.")
	// items := InMemoryCache.Items()
	// for k, v := range items {
	// 	InMemoryCache.Delete(k)
	// 	log.Printf("Cache cleared: Key=%v, Value=%v", k, v.Object)
	// }
	// log.Println("Cache cleared!")
}
