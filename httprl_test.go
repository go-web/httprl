package httprl

import (
	"context"
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

// dualBackend implements both Backend and BackendContext. The wiring test
// asserts that RateLimiter prefers HitContext and forwards the request
// context.
type dualBackend struct {
	hitCalled, hitCtxCalled int
	gotCtx                  context.Context
}

func (b *dualBackend) Hit(key string, ttlsec int32) (uint64, int32, error) {
	b.hitCalled++
	return 1, ttlsec, nil
}

func (b *dualBackend) HitContext(ctx context.Context, key string, ttlsec int32) (uint64, int32, error) {
	b.hitCtxCalled++
	b.gotCtx = ctx
	return 1, ttlsec, nil
}

func TestRateLimiterPrefersBackendContext(t *testing.T) {
	b := &dualBackend{}
	rl := &RateLimiter{
		Backend:  b,
		Limit:    10,
		Interval: 1,
		KeyMaker: func(r *http.Request) string { return "k" },
	}
	type ctxKey struct{}
	h := rl.HandleFunc(func(w http.ResponseWriter, r *http.Request) {})
	r := httptest.NewRequest("GET", "/", nil)
	r = r.WithContext(context.WithValue(context.Background(), ctxKey{}, "marker"))
	h(httptest.NewRecorder(), r)
	if b.hitCalled != 0 {
		t.Fatalf("legacy Hit called %d times, want 0", b.hitCalled)
	}
	if b.hitCtxCalled != 1 {
		t.Fatalf("HitContext called %d times, want 1", b.hitCtxCalled)
	}
	if b.gotCtx == nil || b.gotCtx.Value(ctxKey{}) != "marker" {
		t.Fatal("HitContext did not receive the request context")
	}
}

// legacyBackend implements only Backend. The wiring test asserts that
// RateLimiter falls back to Hit when the backend has no HitContext.
type legacyBackend struct{ hitCalled int }

func (b *legacyBackend) Hit(key string, ttlsec int32) (uint64, int32, error) {
	b.hitCalled++
	return 1, ttlsec, nil
}

func TestRateLimiterFallsBackToHit(t *testing.T) {
	b := &legacyBackend{}
	rl := &RateLimiter{
		Backend:  b,
		Limit:    10,
		Interval: 1,
		KeyMaker: func(r *http.Request) string { return "k" },
	}
	h := rl.HandleFunc(func(w http.ResponseWriter, r *http.Request) {})
	r := httptest.NewRequest("GET", "/", nil)
	h(httptest.NewRecorder(), r)
	if b.hitCalled != 1 {
		t.Fatalf("Hit called %d times, want 1", b.hitCalled)
	}
}
