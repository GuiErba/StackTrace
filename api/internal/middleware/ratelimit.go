package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"stacktrace/internal/cache"
)

const (
	rateLimitMax    = 10000
	rateLimitWindow = 60 * time.Second
)

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID, ok := GetProjectID(c)
		if !ok {
			c.Next()
			return
		}

		key := fmt.Sprintf("stacktrace:ratelimit:%s", projectID.String())

		count, err := cache.Client.Incr(cache.Ctx, key).Result()
		if err != nil {
			log.Printf("Redis error on rate limit INCR: %v", err)
			c.Next()
			return
		}

		if count == 1 {
			cache.Client.Expire(cache.Ctx, key, rateLimitWindow)
		}

		remaining := rateLimitMax - int(count)
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", strconv.Itoa(rateLimitMax))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

		if count > int64(rateLimitMax) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"limit":       rateLimitMax,
				"retry_after": "60s",
			})
			return
		}

		c.Next()
	}
}
