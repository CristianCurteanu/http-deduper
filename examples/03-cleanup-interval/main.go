package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/CristianCurteanu/http-deduper/cache"
)

func main() {
	cli := 3 * time.Second
	fmt.Printf("cli: %+v\n", cli)
	client := cache.NewCache(30*time.Second, cache.WithCleanupInterval(cli))
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// This is where default TTL is overriden
	_, err := client.Fetch(ctx, "https://google.com", time.Minute)
	if err != nil {
		log.Fatalf("fetch failed, err=%q", err)
	}

	// log.Printf("data: %s", string(data))

	hits, misses, entries := client.Stats()
	log.Printf("\nStatistics:\n\tHits: %d\n\tMisses: %d\n\tEntries: %d\n", hits, misses, entries)

	time.Sleep(6 * time.Second)
	hits, misses, entries = client.Stats()
	log.Printf("\nStatistics:\n\tHits: %d\n\tMisses: %d\n\tEntries: %d\n", hits, misses, entries)
}
