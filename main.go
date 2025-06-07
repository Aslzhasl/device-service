// main.go

package main

import (
	"context"
	"fmt"
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
	"github.com/redis/go-redis/v9"
)

func main() {
	// Загрузка .env (если есть)
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, relying on real env vars")
	}

	// 1) Проверяем и инициализируем Redis
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
	log.Println("✅ Redis connected via URL")

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

	// 3) Инициализируем Firebase Storage (и Redis внутри config.InitRedis, если нужно)
	config.InitRedis() // оставляем, чтобы config.RedisClient был готов

	// 4) Создаём репозитории
	deviceRepo := repository.NewDeviceRepository(db)
	favRepo := repository.NewFavoriteRepository(db)

	// 5) Настраиваем Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// 6) Группа /api + JWT middleware
	api := router.Group("/api")
	api.Use(middleware.JWTAuthMiddleware())

	// 7) Регистрируем маршруты в нужном порядке

	// 7.1. Загрузка файлов: POST /api/upload (multipart/form-data)
	handler.RegisterUploadHandler(api)

	// 7.2. CRUD для устройств: POST/GET/PUT/DELETE /api/devices
	handler.RegisterDeviceRoutes(api, deviceRepo)

	// 7.3. Маршруты для избранного (favorites)
	handler.RegisterFavoriteRoutes(api, favRepo)

	// 7.4. Метаданные (категории, города, регионы, тренды)
	handler.RegisterMetaRoutes(api, deviceRepo)

	// 8) Запуск HTTP-сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Starting server on :%s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
