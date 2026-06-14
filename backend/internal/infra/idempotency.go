package infra

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// SaveIdempotentResponse stores body under key using SET NX (only if not exists).
func SaveIdempotentResponse(ctx context.Context, rdb *redis.Client, key string, body []byte, ttl time.Duration) error {
	return rdb.SetNX(ctx, key, body, ttl).Err()
}

// GetIdempotentResponse retrieves stored response. Returns (nil, false, nil) on miss.
func GetIdempotentResponse(ctx context.Context, rdb *redis.Client, key string) ([]byte, bool, error) {
	val, err := rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return val, true, nil
}
