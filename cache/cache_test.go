package cache

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestCacheFetch(t *testing.T) {
	client := NewCache(3 * time.Minute)
	defer func() {
		time.Sleep(time.Second)
		client.Close()
	}()

	data, err := client.Fetch(context.Background(), "https://google.com")
	noError(t, err)
	notEmpty(t, data)
	_, _, entries := client.Stats()
	equal(t, entries, 1)

	data, err = client.Fetch(context.Background(), "https://google.com")
	noError(t, err)
	notEmpty(t, data)
	_, _, entries = client.Stats()
	equal(t, entries, 1)
}

func TestCacheConcurrency(t *testing.T) {
	client := NewCache(3 * time.Minute)
	defer func() {
		time.Sleep(time.Second)
		client.Close()
	}()

	urls := []string{
		"https://google.com",
		"https://facebook.com",
		"https://google.com",
		"https://google.com",
		"https://facebook.com",
		"https://facebook.com",
		"https://www.enclaive.io/",
		"https://google.com",
		"https://www.enclaive.io/",
		"https://google.com",
	}
	var w sync.WaitGroup

	w.Add(len(urls))
	for _, u := range urls {

		go func(wg *sync.WaitGroup, url string) {
			defer wg.Done()

			data, err := client.Fetch(context.Background(), url)
			noError(t, err)
			notEmpty(t, data)
		}(&w, u)
	}

	w.Wait()
	hits, misses, entries := client.Stats()
	equal(t, entries, 3)
	equal(t, hits, len(urls))
	equal(t, misses, 3) // This is because misses will count whenever misses cache read, eventually on evict the number will differ from entries

}

func TestCacheTTL(t *testing.T) {
	client := NewCache(2*time.Second, WithCleanupInterval(time.Second))
	defer func() {
		time.Sleep(time.Second)
		client.Close()
	}()

	data, err := client.Fetch(context.Background(), "https://google.com")
	noError(t, err)
	notEmpty(t, data)
	_, _, entries := client.Stats()
	equal(t, entries, 1)

	time.Sleep(3 * time.Second)
	_, _, entries = client.Stats()
	equal(t, entries, 0)

}

func noError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("expected nil error, got: %q", err)
	}
}

func notEmpty(t *testing.T, o any) {
	objValue := reflect.ValueOf(o)

	zero := reflect.Zero(objValue.Type())
	if reflect.DeepEqual(o, zero.Interface()) {
		t.Fatalf("the value received is empty: %+v", o)
	}
}

func equal(t *testing.T, o1 any, o2 any) {
	if !reflect.DeepEqual(o1, o2) {
		t.Fatalf("the value %+v, is not equal to %+v", o1, o2)
	}
}
