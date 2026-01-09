package config

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis() *redis.Client {
	var ctx = context.Background()
	redisUrl := LoadEnv().RedisUrl

	connect := redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	if err := connect.Ping(ctx).Err(); err != nil {
		panic("❌ Failed to connect to Redis: " + err.Error())
	} else {
		log.Println("✅ Connected to Redis")
	}
	return connect
}
