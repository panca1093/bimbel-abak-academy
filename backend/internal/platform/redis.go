package platform

import "github.com/redis/go-redis/v9"

func NewRedis(addr, password string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})
}
