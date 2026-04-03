package middleware

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"stacktrace/internal/cache"
	"stacktrace/internal/repository"
)

const apiKeyCacheTTL = 5 * time.Minute

func Auth(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header must be: Bearer {api_key}",
			})
			return
		}

		apiKey := parts[1]
		cacheKey := fmt.Sprintf("stacktrace:apikey:%s", hashAPIKey(apiKey))

		projectIDStr, err := cache.Client.Get(cache.Ctx, cacheKey).Result()
		if err == nil {
			projectID, err := uuid.Parse(projectIDStr)
			if err == nil {
				c.Set("project_id", projectID)
				c.Next()
				return
			}
		} else if err != redis.Nil {
			log.Printf("Redis error on auth cache lookup: %v", err)
		}

		project, err := repository.GetProjectByAPIKey(db, apiKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			return
		}

		err = cache.Client.Set(cache.Ctx, cacheKey, project.ID.String(), apiKeyCacheTTL).Err()
		if err != nil {
			log.Printf("Redis error on auth cache set: %v", err)
		}

		c.Set("project_id", project.ID)
		c.Next()
	}
}

func GetProjectID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("project_id")
	if !exists {
		return uuid.UUID{}, false
	}
	projectID, ok := val.(uuid.UUID)
	return projectID, ok
}

func hashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return fmt.Sprintf("%x", hash)
}
