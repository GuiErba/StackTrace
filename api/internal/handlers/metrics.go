package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"stacktrace/internal/middleware"
	"stacktrace/internal/repository"
)

type MetricsHandler struct {
	DB *sql.DB
}

func NewMetricsHandler(db *sql.DB) *MetricsHandler {
	return &MetricsHandler{DB: db}
}

func (h *MetricsHandler) Overview(c *gin.Context) {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id not found in context"})
		return
	}

	metrics, err := repository.GetOverviewMetrics(h.DB, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}
