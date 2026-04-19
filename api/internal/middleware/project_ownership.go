package middleware

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"stacktrace/internal/repository"
)

func ProjectOwnership(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		projectIDStr := c.Query("project_id")
		if projectIDStr == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "project_id query parameter required"})
			return
		}

		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid project_id"})
			return
		}

		_, err = repository.GetProjectByIDAndUser(db, projectID, userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}

		c.Set("project_id", projectID)
		c.Next()
	}
}
