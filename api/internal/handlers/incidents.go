package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"stacktrace/internal/middleware"
	"stacktrace/internal/models"
	"stacktrace/internal/repository"
)

type IncidentHandler struct {
	DB *sql.DB
}

func NewIncidentHandler(db *sql.DB) *IncidentHandler {
	return &IncidentHandler{DB: db}
}

func getProjectIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	projectID, ok := middleware.GetProjectID(c)
	if ok {
		return projectID, true
	}

	projectIDStr := c.Query("project_id")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id required"})
		return uuid.Nil, false
	}

	parsed, err := uuid.Parse(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project_id"})
		return uuid.Nil, false
	}

	return parsed, true
}

func (h *IncidentHandler) List(c *gin.Context) {
	projectID, ok := getProjectIDFromContext(c)
	if !ok {
		return
	}

	incidents, err := repository.ListIncidents(h.DB, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list incidents"})
		return
	}

	if incidents == nil {
		incidents = []models.Incident{}
	}

	c.JSON(http.StatusOK, gin.H{"data": incidents})
}

func (h *IncidentHandler) Resolve(c *gin.Context) {
	projectID, ok := getProjectIDFromContext(c)
	if !ok {
		return
	}

	incidentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	if err := repository.ResolveIncident(h.DB, incidentID, projectID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Incident resolved"})
}
