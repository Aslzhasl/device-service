package config

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func InitRedis() {
	// 1) Подгрузили URL (который у вас уже в REDIS_URL)
	redURL := os.Getenv("REDIS_URL")
	opt, err := redis.ParseURL(redURL)
	if err != nil {
		log.Fatalf("🔴 redis parse url failed: %v", err)
	}

	// 2) Создали клиента по этому URL (с TLS автоматически)
	RedisClient = redis.NewClient(opt)

	// 3) Проверили соединение
	if err := RedisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("🔴 redis connection failed: %v", err)
	}

	log.Println("✅ Redis connected via URL")
}
