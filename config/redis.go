package config

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
)

var RedisClient *redis.Client

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"), // e.g. localhost:6379
		Password: "",                      // no password by default
		DB:       0,
	})

	if err := RedisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("❌ Redis connection failed: %v", err)
	}

	log.Println("✅ Connected to Redis")
}
