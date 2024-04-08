package db

import (
	"github.com/redis/go-redis/v9"
	"context"
)

var Redis *redis.Client


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

func ClearRedis() error {
	ctx := context.Background()
	return Redis.FlushAll(ctx).Err()
}

func DeleteCache() error {
    ctx := context.Background()
	return Redis.Del(ctx, "cacheStringParams").Err()
}