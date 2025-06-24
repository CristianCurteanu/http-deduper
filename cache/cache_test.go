package cache

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestCacheFetch(t *testing.T) {

	requestCount := 0
	expectedResponse := "Hello from test server!"

	server := newTestServer(
		mapping{
			"/test",
			func(w http.ResponseWriter, r *http.Request) {
				requestCount++
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(expectedResponse))
			},
		},
	)
	defer server.Close()

	client := NewCache(3 * time.Minute)
	defer func() {
		time.Sleep(time.Second)
		client.Close()
	}()

	data, err := client.Fetch(context.Background(), server.URL+"/test")
	noError(t, err)
	notEmpty(t, data)
	equal(t, string(data), expectedResponse)
	_, _, entries := client.Stats()
	equal(t, entries, 1)

	data, err = client.Fetch(context.Background(), server.URL+"/test")
	noError(t, err)
	notEmpty(t, data)
	equal(t, string(data), expectedResponse)
	_, _, entries = client.Stats()
	equal(t, entries, 1)
}

func TestCacheConcurrency(t *testing.T) {
	client := NewCache(3 * time.Minute)
	defer func() {
		time.Sleep(time.Second)
		client.Close()
	}()

	server := newTestServer(
		mapping{
			"/test-1",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("response from test-1 endpoint"))
			},
		},
		mapping{
			"/test-2",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("response from test-2 endpoint"))
			},
		},
		mapping{
			"/test-3",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("response from test-3 endpoint"))
			},
		},
	)
	defer server.Close()

	urls := []string{
		server.URL + "/test-1",
		server.URL + "/test-2",
		server.URL + "/test-1",
		server.URL + "/test-1",
		server.URL + "/test-2",
		server.URL + "/test-2",
		server.URL + "/test-3",
		server.URL + "/test-1",
		server.URL + "/test-3",
		server.URL + "/test-1",
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
	equal(t, hits, 0)
	equal(t, misses, 10)

}

func TestCacheTTL(t *testing.T) {
	client := NewCache(2*time.Second, WithCleanupInterval(time.Second))
	defer func() {
		time.Sleep(time.Second)
		client.Close()
	}()

	server := newTestServer(
		mapping{
			"/test",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("test"))
			},
		},
	)
	defer server.Close()

	data, err := client.Fetch(context.Background(), server.URL+"/test")
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

type mapping struct {
	route   string
	handler func(http.ResponseWriter, *http.Request)
}

func newTestServer(routes ...mapping) *httptest.Server {
	mux := http.NewServeMux()

	for _, r := range routes {
		mux.HandleFunc(r.route, r.handler)
	}

	return httptest.NewServer(mux)
}
