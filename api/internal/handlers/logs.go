package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"stacktrace/internal/middleware"
	"stacktrace/internal/models"
	"stacktrace/internal/repository"
	"stacktrace/internal/services"
)

type LogHandler struct {
	DB *sql.DB
}

func NewLogHandler(db *sql.DB) *LogHandler {
	return &LogHandler{DB: db}
}

func (h *LogHandler) IngestLog(c *gin.Context) {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "project_id not found in context"})
		return
	}

	var input models.LogInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	timestamp := time.Now().UTC()
	if input.Timestamp != nil {
		timestamp = *input.Timestamp
	}

	service := "default"
	if input.Service != "" {
		service = input.Service
	}

	entry := models.LogEntry{
		ProjectID: projectID,
		Timestamp: timestamp,
		Level:     input.Level,
		Message:   input.Message,
		Service:   service,
		TraceID:   input.TraceID,
		Metadata:  input.Metadata,
	}

	if !services.Enqueue(entry) {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Server is overloaded, please retry later",
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status":  "accepted",
		"message": "Log queued for processing",
	})
}

func (h *LogHandler) IngestBatch(c *gin.Context) {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "project_id not found in context"})
		return
	}

	var inputs []models.LogInput
	if err := c.ShouldBindJSON(&inputs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(inputs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty batch"})
		return
	}

	if len(inputs) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch too large, max 500 logs per request"})
		return
	}

	accepted := 0
	dropped := 0

	for _, input := range inputs {
		timestamp := time.Now().UTC()
		if input.Timestamp != nil {
			timestamp = *input.Timestamp
		}

		service := "default"
		if input.Service != "" {
			service = input.Service
		}

		entry := models.LogEntry{
			ProjectID: projectID,
			Timestamp: timestamp,
			Level:     input.Level,
			Message:   input.Message,
			Service:   service,
			TraceID:   input.TraceID,
			Metadata:  input.Metadata,
		}

		if services.Enqueue(entry) {
			accepted++
		} else {
			dropped++
		}
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status":   "accepted",
		"accepted": accepted,
		"dropped":  dropped,
	})
}

func (h *LogHandler) QueryLogs(c *gin.Context) {
	projectID, ok := middleware.GetProjectID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "project_id not found in context"})
		return
	}

	filters := models.LogFilter{
		Level:   c.Query("level"),
		Service: c.Query("service"),
		TraceID: c.Query("trace_id"),
		Cursor:  c.Query("cursor"),
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 {
			filters.Limit = limit
		}
	}

	if fromStr := c.Query("from"); fromStr != "" {
		from, err := time.Parse(time.RFC3339, fromStr)
		if err == nil {
			filters.From = &from
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		to, err := time.Parse(time.RFC3339, toStr)
		if err == nil {
			filters.To = &to
		}
	}

	logs, nextCursor, hasMore, err := repository.QueryLogs(h.DB, projectID, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        logs,
		"next_cursor": nextCursor,
		"has_more":    hasMore,
	})
}
