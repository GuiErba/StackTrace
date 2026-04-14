package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"stacktrace/internal/middleware"
	"stacktrace/internal/repository"
)

type ProjectHandler struct {
	DB *sql.DB
}

func NewProjectHandler(db *sql.DB) *ProjectHandler {
	return &ProjectHandler{DB: db}
}

type CreateProjectInput struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
	Slug string `json:"slug" binding:"required,min=2,max=50"`
}

func generateAPIKey() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return "sk_live_" + hex.EncodeToString(b)
}

func hashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

func (h *ProjectHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var input CreateProjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiKey := generateAPIKey()
	apiKeyHash := hashAPIKey(apiKey)
	apiKeyPrefix := apiKey[:16]

	project, err := repository.CreateProjectWithUser(h.DB, input.Name, input.Slug, userID, apiKeyHash, apiKeyPrefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create project: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      project.ID,
		"name":    project.Name,
		"slug":    project.Slug,
		"api_key": apiKey,
		"message": "Save this API key now. You will not be able to see it again.",
	})
}

func (h *ProjectHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	projects, err := repository.GetProjectsByUserID(h.DB, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list projects"})
		return
	}

	type projectResponse struct {
		ID           uuid.UUID `json:"id"`
		Name         string    `json:"name"`
		Slug         *string   `json:"slug,omitempty"`
		APIKeyPrefix *string   `json:"api_key_prefix,omitempty"`
		OwnerEmail   string    `json:"owner_email"`
	}

	var result []projectResponse
	for _, p := range projects {
		result = append(result, projectResponse{
			ID:           p.ID,
			Name:         p.Name,
			Slug:         p.Slug,
			APIKeyPrefix: p.APIKeyPrefix,
			OwnerEmail:   p.OwnerEmail,
		})
	}

	if result == nil {
		result = []projectResponse{}
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *ProjectHandler) RotateKey(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	newKey := generateAPIKey()
	newHash := hashAPIKey(newKey)
	newPrefix := newKey[:16]

	err = repository.RotateProjectAPIKey(h.DB, projectID, userID, newHash, newPrefix)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"api_key": newKey,
		"message": "API key rotated. Save this new key now. The old key has been invalidated.",
	})
}
