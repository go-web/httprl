package memcacherl

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

func TestClient(t *testing.T) {
	addr := os.Getenv("MEMCACHE_ADDR")
	if addr == "" {
		addr = "localhost:11211"
	}
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Skipf("memcache not available at %s: %v", addr, err)
	}
	conn.Close()
	mc := memcache.New(addr)
	c := New(mc)
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
