package main

import (
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
	// 1) Load environment variables from .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, relying on real env vars")
	}

	// 2) Connect to Postgres
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL env var is required")
	}
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Postgres: %v", err)
	}

	// 3) Initialize Redis
	config.InitRedis()

	// 4) Initialize Firebase (Storage)
	config.InitFirebase()

	// 5) Create repositories
	deviceRepo := repository.NewDeviceRepository(db)
	favRepo := repository.NewFavoriteRepository(db)

	// 6) Set up Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 7) Mount your API group with JWT auth
	api := r.Group("/api")
	api.Use(middleware.JWTAuthMiddleware())

	// 8) Register routes
	handler.RegisterDeviceRoutes(api, deviceRepo)
	handler.RegisterFavoriteRoutes(api, favRepo)
	handler.RegisterUploadURLRoute(api) // if you added upload-url
	handler.RegisterMetaRoutes(api, deviceRepo)

	// 9) Start HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // default port
	}
	log.Printf("🚀 Starting server on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
