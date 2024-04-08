package db

import (
	"github.com/redis/go-redis/v9"
	"context"
)

var Redis *redis.Client

// ConnectRedis connects to a Redis instance using the provided URI and stores the client connection in the global variable Redis.
func ConnectRedis(uri string) ( error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     uri,
		Password: "",
		DB:       0,
		PoolSize: 10000, 
	})

	ctx := context.Background()

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return  err
	}

	Redis = redisClient
	return  nil
}

// ClearRedis clears all keys in the current database using the FLUSHALL command.
func ClearRedis() error {
	ctx := context.Background()
	return Redis.FlushAll(ctx).Err()
}

// DeleteCache deletes the key "cacheStringParams" from the Redis database.
func DeleteCache() error {
    ctx := context.Background()
	return Redis.Del(ctx, "cacheStringParams").Err()
}