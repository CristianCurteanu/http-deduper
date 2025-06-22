package cache

import (
	"context"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type cachedValue struct {
	val       []byte
	createdAt time.Time
	expiresAt time.Time
}

func (c *cachedValue) isExpired() bool {
	if c != nil {
		return false
	}
	return time.Now().UTC().After(c.expiresAt)
}

type cacheStats struct {
	hits   int
	misses int
}

type Cache struct {
	// TODO: implement cache storage, TTLs, locking, and deduplication
	mx         *sync.RWMutex
	storage    map[string]*cachedValue
	ttl        time.Duration
	cleanup    time.Duration
	maxSize    uint64
	closeEvict chan struct{}
	stats      *cacheStats
}

type cacheOption func(*Cache)

func NewCache(defaultTTL time.Duration, opts ...cacheOption) *Cache {
	c := &Cache{
		mx:         &sync.RWMutex{},
		ttl:        defaultTTL,
		cleanup:    time.Minute,
		maxSize:    5000,
		stats:      &cacheStats{},
		storage:    make(map[string]*cachedValue, 5000),
		closeEvict: make(chan struct{}),
	}

	for _, optFunc := range opts {
		optFunc(c)
	}

	go c.startCleanup()

	return c
}

func (c *Cache) Fetch(ctx context.Context, url string, ttlOverride ...time.Duration) ([]byte, error) {
	ttl := c.ttl
	if len(ttlOverride) >= 1 {
		ttl = ttlOverride[0]
	}

	c.mx.Lock()
	defer c.mx.Unlock()

	c.stats.hits++
	val, exists := c.storage[url]
	if exists && !val.isExpired() {
		return val.val, nil
	}

	c.stats.misses++
	respBytes, err := makeHttpReq(url)
	if err != nil {
		return nil, err
	}

	timeNow := time.Now().UTC()
	cacheVal := cachedValue{
		val:       respBytes,
		expiresAt: timeNow.Add(ttl),
		createdAt: timeNow,
	}
	c.storage[url] = &cacheVal

	return respBytes, nil
}

func (c *Cache) Stats() (hits int, misses int, entries int) {
	c.mx.Lock()
	defer c.mx.Unlock()

	return c.stats.hits, c.stats.misses, len(c.storage)
}

func (c *Cache) Close() {
	c.closeEvict <- struct{}{}
}

func (c *Cache) startCleanup() {
	ticker := time.NewTicker(c.cleanup)

	running := true
	for running {
		select {
		case <-c.closeEvict:
			running = false
			close(c.closeEvict)
			ticker.Stop()
		case <-ticker.C:
			c.evictExpired()
		default:
			log.Println("")
		}
	}
}

func (c *Cache) evictExpired() {
	c.mx.Lock()
	defer c.mx.Unlock()

	now := time.Now()

	for key, val := range c.storage {
		if now.After(val.expiresAt) {
			delete(c.storage, key)
		}
	}
}

func makeHttpReq(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/html")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func WithMaxSize(maxSize uint64) cacheOption {
	return func(c *Cache) {
		c.maxSize = maxSize
	}
}

func WithCleanupInterval(cd time.Duration) cacheOption {
	return func(c *Cache) {
		c.cleanup = cd
	}
}
