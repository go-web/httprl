package httprl_test

import (
	"io"
	"net/http"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/fiorix/go-redis/redis"

	"github.com/go-web/httprl"
	"github.com/go-web/httprl/memcacherl"
	"github.com/go-web/httprl/redisrl"
)

func myHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello, world")
}

func ExampleMap() {
	rl := &httprl.RateLimiter{
		Backend:  httprl.NewMap(1),
		Limit:    5,
		Interval: 1,
		KeyMaker: func(r *http.Request) string {
			return r.Header.Get("X-Auth-Token")
		},
	}
	mux := http.NewServeMux()
	mux.Handle("/", rl.HandleFunc(myHandler))
	http.ListenAndServe(":8080", mux)
}

func ExampleMemcache() {
	mc := memcache.New("localhost:11211")
	rl := &httprl.RateLimiter{
		Backend:  memcacherl.New(mc),
		Limit:    5,
		Interval: 1,
	}
	mux := http.NewServeMux()
	mux.Handle("/", rl.HandleFunc(myHandler))
	http.ListenAndServe(":8080", mux)
}

func ExampleRedis() {
	rc := redis.New("localhost:6379")
	rl := &httprl.RateLimiter{
		Backend:  redisrl.New(rc),
		Limit:    5,
		Interval: 1,
	}
	mux := http.NewServeMux()
	mux.Handle("/", rl.HandleFunc(myHandler))
	http.ListenAndServe(":8080", mux)
}
