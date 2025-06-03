package main

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"os"

	"device-service/config"
	"device-service/internal/handler"
	"device-service/internal/middleware"
	"device-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, relying on real env vars")
	}

	// 1) Проверяем REDIS_URL
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		log.Fatal("REDIS_URL env var is required")
	}
	fmt.Println("REDIS_URL:", redisUrl)

	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		log.Fatalf("🙈 redis.ParseURL failed: %v", err)
	}
	client := redis.NewClient(opt)

	ctx := context.Background()
	if err := client.Set(ctx, "testkey", "testvalue", 0).Err(); err != nil {
		log.Fatalf("❌ Redis SET failed: %v", err)
	}

	// 2) Подключаемся к Postgres
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL env var is required")
	}
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Postgres: %v", err)
	}
	log.Println("✅ Connected to Postgres")

	// 3) Инициализируем Redis и Firebase
	config.InitRedis()
	config.InitFirebase()

	// 4) Создаем репозитории/сервисы/хендлеры
	deviceRepo := repository.NewDeviceRepository(db)
	favRepo := repository.NewFavoriteRepository(db)

	// 5) Настраиваем Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	api := r.Group("/api")
	api.Use(middleware.JWTAuthMiddleware())

	handler.RegisterDeviceRoutes(api, deviceRepo)
	handler.RegisterFavoriteRoutes(api, favRepo)
	handler.RegisterUploadURLRoute(api)
	handler.RegisterMetaRoutes(api, deviceRepo)

	// 6) Запуск HTTP-сервера на PORT (по умолчанию 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Starting server on :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
