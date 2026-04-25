package redisrl

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func dialRedis(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Skipf("redis not available at %s: %v", addr, err)
	}
	conn.Close()
	return redis.NewClient(&redis.Options{Addr: addr})
}

func TestClient(t *testing.T) {
	rc := dialRedis(t)
	defer rc.Close()
	c := New(rc)
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
	rc := dialRedis(t)
	defer rc.Close()
	c := New(rc)
	// Drive the HitContext path explicitly.
	n, _, err := c.HitContext(context.Background(), "hello-ctx", 1)
	if err != nil {
		t.Fatal(err)
	}
	if n == 0 {
		t.Fatal("zero count on first hit")
	}
	// Cancelled context should fail before talking to redis.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := c.HitContext(ctx, "hello-ctx", 1); err == nil {
		t.Fatal("expected error on cancelled context")
	}
}
