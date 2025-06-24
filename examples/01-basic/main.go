package main

import (
	"context"
	"log"
	"time"

	"github.com/CristianCurteanu/http-deduper/cache"
)

func main() {
	client := cache.NewCache(30 * time.Second)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	data, err := client.Fetch(ctx, "https://google.com")
	if err != nil {
		log.Fatalf("fetch failed, err=%q", err)
	}

	log.Printf("data: %s", string(data))

	hits, misses, entries := client.Stats()
	log.Printf("\nStatistics:\n\tHits: %d\n\tMisses: %d\n\tEntries: %d\n", hits, misses, entries)
}
