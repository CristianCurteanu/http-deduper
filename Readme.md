# http-deduper

This repository provides a HTTP request deduplication client, by storing requests with a defined time to live (TTL).

## Installation

In order to make use of this package, you need to install it as a module to your project:

```sh
go get github.com/CristianCurteanu/http-deduper
```

Then, make sure it was installed to your `go.mod` file

## Usage

Once, installed you need to access `cache` subpackage, and create a client instance, using `NewCache` function:

```go
client := cache.NewCache(3 * time.Minute)
defer client.Close()

// Without TTL override
data, err := client.Fetch(context.Background(), "https://google.com")
if err != nil {
    // handle error
}

// With TTL overried
data, err := client.Fetch(context.Background(), "https://linkedin.com", 10*time.Second)
if err != nil {
    // handle error
}

// handle `data` as []byte

hits, misses, entries := client.Stats()
fmt.Printf("Hits: %d\nMisses: %d\nEntries: %d", hits, misses, entries)
```

The `defaultTTL` parameter for `NewCache` functions needs to be defined as a fallback value for TTL.

Also, there is possibility to override `defaultTTL` by passing an additional parameter to `Fetch` method, which will override the default value of the TTL. As in example above, the request to `https://facebook.com` will be cached only for 10 seconds.

### Override cleanup cycles

The cached responses are usually cleaned up automatically once TTL expire. By default, a cleanup cycle happens every minute after client initialization, but this interval could be defined as a different value, here is how to do it:

```go
client := cache.NewCache(3 * time.Minute, cache.WithCleanupInterval(30 * time.Second))
defer client.Close()
```

By passing `WithCleanupInterval` value to the `NewCache` function, the cleanup interval changed to 30 seconds

## Known issues

There is though space for improvement:

- [ ] Make the `Fetch` support other HTTP verbs, as currently it does support only `GET`
- [ ] Make it possible to send request body and headers
- [ ] Add serialization/deserialization to request/response bodies
- [ ] Handle status codes