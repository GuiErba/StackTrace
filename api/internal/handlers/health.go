package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"stacktrace/internal/cache"
)

type HealthHandler struct {
	DB *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{DB: db}
}

// Check is a lightweight health check that does NOT ping the database.
// This allows Neon's compute to auto-suspend and save CU-hours.
// Used by BetterStack and uptime monitors.
func (h *HealthHandler) Check(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// DeepCheck pings the database and Redis for full diagnostics.
// Use manually when you need to verify all dependencies are healthy.
func (h *HealthHandler) DeepCheck(c *gin.Context) {
	status := "ok"
	httpStatus := http.StatusOK

	dbStatus := "ok"
	if err := h.DB.Ping(); err != nil {
		dbStatus = "error"
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	redisStatus := "ok"
	if err := cache.Client.Ping(cache.Ctx).Err(); err != nil {
		redisStatus = "error"
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"status":    status,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"database":  dbStatus,
		"redis":     redisStatus,
	})
}
