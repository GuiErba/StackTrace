package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"stacktrace/internal/cache"
	"stacktrace/internal/repository"
)

type StatusHandler struct {
	DB *sql.DB
}

func NewStatusHandler(db *sql.DB) *StatusHandler {
	return &StatusHandler{DB: db}
}

type StatusResponse struct {
	ProjectName string          `json:"project_name"`
	Status      string          `json:"status"`
	UpdatedAt   string          `json:"updated_at"`
	Incidents   []IncidentBrief `json:"incidents"`
}

type IncidentBrief struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Status      string  `json:"status"`
	StartedAt   string  `json:"started_at"`
	ResolvedAt  *string `json:"resolved_at,omitempty"`
}

const statusCacheTTL = 30 * time.Second

func (h *StatusHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	cacheKey := "stacktrace:status:" + slug
	cached, err := cache.Client.Get(cache.Ctx, cacheKey).Result()
	if err == nil {
		var response StatusResponse
		if json.Unmarshal([]byte(cached), &response) == nil {
			c.JSON(http.StatusOK, response)
			return
		}
	}

	project, err := repository.GetProjectBySlug(h.DB, slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	incidents, err := repository.ListRecentIncidents(h.DB, project.ID, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load status"})
		return
	}

	status := "operational"
	var briefs []IncidentBrief

	for _, inc := range incidents {
		if inc.Status == "open" {
			status = "incident"
		}

		brief := IncidentBrief{
			ID:          inc.ID.String(),
			Title:       inc.Title,
			Description: inc.Description,
			Status:      inc.Status,
			StartedAt:   inc.StartedAt.Format(time.RFC3339),
		}

		if inc.ResolvedAt != nil {
			resolved := inc.ResolvedAt.Format(time.RFC3339)
			brief.ResolvedAt = &resolved
		}

		briefs = append(briefs, brief)
	}

	if briefs == nil {
		briefs = []IncidentBrief{}
	}

	response := StatusResponse{
		ProjectName: project.Name,
		Status:      status,
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
		Incidents:   briefs,
	}

	jsonBytes, err := json.Marshal(response)
	if err == nil {
		cache.Client.Set(cache.Ctx, cacheKey, string(jsonBytes), statusCacheTTL)
	} else {
		log.Printf("Failed to cache status page: %v", err)
	}

	c.JSON(http.StatusOK, response)
}
