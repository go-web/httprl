package httprl

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
)

func TestRateLimiter(t *testing.T) {
	counter := struct {
		sync.Mutex
		n int
	}{}
	f := func(w http.ResponseWriter, r *http.Request) {
		counter.Lock()
		counter.n++
		counter.Unlock()
	}
	m := NewMap(1)
	m.Start()
	defer m.Stop()
	rl := &RateLimiter{
		Backend:  m,
		Limit:    2,
		Interval: 1,
		KeyMaker: func(r *http.Request) string {
			return "rate-limiter-test"
		},
	}
	mux := http.NewServeMux()
	mux.Handle("/", rl.Handle(http.HandlerFunc(f)))
	s := httptest.NewServer(mux)
	defer s.Close()
	for i := 0; i < 3; i++ {
		resp, err := http.Get(s.URL)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusServiceUnavailable {
			t.Skip("Backend unavailable, cannot proceed")
		}
		lim, _ := strconv.Atoi(resp.Header.Get("X-RateLimit-Limit"))
		rem, _ := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining"))
		res, _ := strconv.Atoi(resp.Header.Get("X-RateLimit-Reset"))
		switch {
		case i == 0 && lim == 2 && rem == 1 && res > 0:
		case (i == 1 || i == 2) && lim == 2 && rem == 0 && res == 0:
		default:
			t.Fatalf("Test %d: Unexpected values: limit=%d, remaining=%d, reset=%d",
				i, lim, rem, res)
		}
	}
}
