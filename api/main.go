package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"stacktrace/internal/cache"
	"stacktrace/internal/database"
	"stacktrace/internal/handlers"
	"stacktrace/internal/middleware"
	"stacktrace/internal/services"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	db, err := database.Connect(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	defer db.Close()
	log.Println("Database connected")

	cache.Init()
	if err := cache.Client.Ping(cache.Ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis: ", err)
	}
	log.Println("Redis connected")

	services.InitIngest(db)
	log.Println("Ingest workers started")

	router := gin.Default()

	healthHandler := handlers.NewHealthHandler(db)
	logHandler := handlers.NewLogHandler(db)

	router.GET("/health", healthHandler.Check)

	auth := router.Group("/")
	auth.Use(middleware.Auth(db))
	auth.Use(middleware.RateLimit())
	{
		auth.POST("/logs", logHandler.IngestLog)
		auth.POST("/logs/batch", logHandler.IngestBatch)
		auth.GET("/logs", logHandler.QueryLogs)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("StackTrace API running on http://localhost:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
