package main

import (
	"device-service/config"
	"device-service/internal/handler"
	"device-service/internal/middleware"
	"device-service/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"os"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("❌ Error loading .env file")
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("❌ DATABASE_URL not set")
	}

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}

	r := gin.Default()
	config.InitRedis()
	repo := repository.NewDeviceRepository(db)

	api := r.Group("/api")
	api.Use(middleware.JWTAuthMiddleware())

	handler.RegisterDeviceRoutes(api, repo)

	if err := r.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
