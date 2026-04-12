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
	"stacktrace/pkg/notify"
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

	var emailNotifier *notify.EmailNotifier
	resendKey := os.Getenv("RESEND_API_KEY")
	resendFrom := os.Getenv("RESEND_FROM")
	if resendKey != "" && resendFrom != "" {
		emailNotifier = notify.NewEmailNotifier(resendKey, resendFrom)
		log.Println("Email notifier initialized")
	} else {
		log.Println("Warning: RESEND_API_KEY or RESEND_FROM not set, email alerts disabled")
	}

	services.InitAlertWorker(db, emailNotifier)
	log.Println("Alert worker started")

	router := gin.Default()

	healthHandler := handlers.NewHealthHandler(db)
	logHandler := handlers.NewLogHandler(db)
	alertRuleHandler := handlers.NewAlertRuleHandler(db)
	incidentHandler := handlers.NewIncidentHandler(db)
	statusHandler := handlers.NewStatusHandler(db)

	router.GET("/health", healthHandler.Check)
	router.GET("/status/:slug", statusHandler.GetBySlug)

	auth := router.Group("/")
	auth.Use(middleware.Auth(db))
	auth.Use(middleware.RateLimit())
	{
		auth.POST("/logs", logHandler.IngestLog)
		auth.POST("/logs/batch", logHandler.IngestBatch)
		auth.GET("/logs", logHandler.QueryLogs)

		auth.POST("/alert-rules", alertRuleHandler.Create)
		auth.GET("/alert-rules", alertRuleHandler.List)
		auth.DELETE("/alert-rules/:id", alertRuleHandler.Delete)

		auth.GET("/incidents", incidentHandler.List)
		auth.PATCH("/incidents/:id/resolve", incidentHandler.Resolve)
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
