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

func (h *HealthHandler) Check(c *gin.Context) {
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
