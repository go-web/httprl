package redisrl

import (
	"testing"
	"time"

	"github.com/fiorix/go-redis/redis"
)

func TestClient(t *testing.T) {
	rc := redis.New("localhost:6379")
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
