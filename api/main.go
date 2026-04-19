package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
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

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"X-RateLimit-Limit", "X-RateLimit-Remaining"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	healthHandler := handlers.NewHealthHandler(db)
	authHandler := handlers.NewAuthHandler(db, emailNotifier)
	logHandler := handlers.NewLogHandler(db)
	alertRuleHandler := handlers.NewAlertRuleHandler(db)
	incidentHandler := handlers.NewIncidentHandler(db)
	statusHandler := handlers.NewStatusHandler(db)
	projectHandler := handlers.NewProjectHandler(db)
	metricsHandler := handlers.NewMetricsHandler(db)

	// Public routes (no auth)
	router.GET("/health", healthHandler.Check)
	router.GET("/status/:slug", statusHandler.GetBySlug)
	router.POST("/auth/send-code", authHandler.SendCode)
	router.POST("/auth/verify-code", authHandler.VerifyCode)

	// JWT auth routes (project management)
	jwt := router.Group("/")
	jwt.Use(middleware.JWTAuth())
	{
		jwt.GET("/projects", projectHandler.List)
		jwt.POST("/projects", projectHandler.Create)
		jwt.POST("/projects/:id/rotate-key", projectHandler.RotateKey)
	}

	// Dashboard routes (JWT + project ownership verification)
	dashboard := router.Group("/dashboard")
	dashboard.Use(middleware.JWTAuth())
	dashboard.Use(middleware.ProjectOwnership(db))
	{
		dashboard.GET("/logs", logHandler.QueryLogs)
		dashboard.GET("/incidents", incidentHandler.List)
		dashboard.PATCH("/incidents/:id/resolve", incidentHandler.Resolve)
		dashboard.GET("/alert-rules", alertRuleHandler.List)
		dashboard.POST("/alert-rules", alertRuleHandler.Create)
		dashboard.DELETE("/alert-rules/:id", alertRuleHandler.Delete)
		dashboard.GET("/metrics/overview", metricsHandler.Overview)
	}

	// API key auth routes (SDK)
	sdk := router.Group("/")
	sdk.Use(middleware.Auth(db))
	sdk.Use(middleware.RateLimit())
	{
		sdk.POST("/logs", logHandler.IngestLog)
		sdk.POST("/logs/batch", logHandler.IngestBatch)
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
