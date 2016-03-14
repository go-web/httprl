package httprl

import (
	"testing"
	"time"
)

func TestMap(t *testing.T) {
	m := NewMap(1)
	m.Start()
	defer m.Stop()
	for i := 0; i < 3; i++ {
		n, _, err := m.Hit("hello", 1)
		if err != nil {
			t.Fatal(err)
		}
		if i < 2 && n == 0 {
			t.Fatalf("Test %d: Zero count", i)
		}
		if i == 2 && n != 1 {
			t.Fatalf("Test %d: Key did not expire? %d hits", i, n)
		}
		time.Sleep(1800 * time.Millisecond)
	}
}
