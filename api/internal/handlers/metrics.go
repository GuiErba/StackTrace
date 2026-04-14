package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"stacktrace/internal/repository"
)

type MetricsHandler struct {
	DB *sql.DB
}

func NewMetricsHandler(db *sql.DB) *MetricsHandler {
	return &MetricsHandler{DB: db}
}

func (h *MetricsHandler) Overview(c *gin.Context) {
	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id query parameter required"})
		return
	}

	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project_id"})
		return
	}

	metrics, err := repository.GetOverviewMetrics(h.DB, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}
