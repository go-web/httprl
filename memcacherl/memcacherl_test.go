package memcacherl

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

func dialMemcache(t *testing.T) *memcache.Client {
	t.Helper()
	addr := os.Getenv("MEMCACHE_ADDR")
	if addr == "" {
		addr = "localhost:11211"
	}
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Skipf("memcache not available at %s: %v", addr, err)
	}
	conn.Close()
	return memcache.New(addr)
}

func TestClient(t *testing.T) {
	c := New(dialMemcache(t))
	for i := 0; i < 3; i++ {
		n, _, err := c.Hit("hello", 1)
		if err != nil {
			t.Fatal(err)
		}
		if i < 2 && n == 0 {
			t.Fatalf("Test %d: Zero count", i)
		}
		if i == 2 && n != 1 {
			t.Fatalf("Test %d: Key did not expire", i)
		}
		time.Sleep(1100 * time.Millisecond)
	}
}

func TestClientContext(t *testing.T) {
	c := New(dialMemcache(t))
	// HitContext with a live context should behave like Hit.
	n, _, err := c.HitContext(context.Background(), "hello-ctx", 1)
	if err != nil {
		t.Fatal(err)
	}
	if n == 0 {
		t.Fatal("zero count on first hit")
	}
	// HitContext must short-circuit when the context is already cancelled.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := c.HitContext(ctx, "hello-ctx", 1); err == nil {
		t.Fatal("expected error on cancelled context")
	}
}
