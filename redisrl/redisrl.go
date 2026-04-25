// Package redisrl is a redis client wrapper for rate limiting.
package redisrl

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client is a redis client wrapper suitable for rate limiting.
type Client struct {
	rc *redis.Client
}

// New creates and initializes a new Client backed by a redis/go-redis
// client built with redis.NewClient.
func New(rc *redis.Client) *Client {
	return &Client{rc}
}

// Hit implements the httprl.Backend interface. It delegates to HitContext
// with a background context.
func (c *Client) Hit(key string, ttlsec int32) (count uint64, remttl int32, err error) {
	return c.HitContext(context.Background(), key, ttlsec)
}

// HitContext implements the httprl.BackendContext interface.
func (c *Client) HitContext(ctx context.Context, key string, ttlsec int32) (count uint64, remttl int32, err error) {
	rem, err := c.rc.TTL(ctx, key).Result()
	if err != nil {
		return 0, 0, err
	}
	if rem <= 0 {
		return 1, ttlsec, c.rc.SetEx(ctx, key, "1", time.Duration(ttlsec)*time.Second).Err()
	}
	n, err := c.rc.Incr(ctx, key).Result()
	if err != nil {
		return 0, 0, err
	}
	return uint64(n), int32(rem / time.Second), nil
}
