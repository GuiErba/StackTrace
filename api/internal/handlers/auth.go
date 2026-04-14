package handlers

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"stacktrace/internal/cache"
	"stacktrace/internal/repository"
	"stacktrace/pkg/notify"
)

type AuthHandler struct {
	DB       *sql.DB
	Notifier *notify.EmailNotifier
}

func NewAuthHandler(db *sql.DB, notifier *notify.EmailNotifier) *AuthHandler {
	return &AuthHandler{DB: db, Notifier: notifier}
}

type SendCodeInput struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyCodeInput struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

func generateOTP() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	code := fmt.Sprintf("%06d", int(b[0])<<16|int(b[1])<<8|int(b[2]))
	if len(code) > 6 {
		code = code[:6]
	}
	return code
}

func (h *AuthHandler) SendCode(c *gin.Context) {
	var input SendCodeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid email is required"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))

	rateLimitKey := fmt.Sprintf("stacktrace:otp:ratelimit:%s", email)
	count, _ := cache.Client.Incr(cache.Ctx, rateLimitKey).Result()
	if count == 1 {
		cache.Client.Expire(cache.Ctx, rateLimitKey, 15*time.Minute)
	}
	if count > 3 {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many attempts, try again later"})
		return
	}

	code := generateOTP()

	otpKey := fmt.Sprintf("stacktrace:otp:%s", email)
	cache.Client.Set(cache.Ctx, otpKey, code, 10*time.Minute)

	if h.Notifier != nil {
		go func() {
			if err := h.Notifier.SendOTP(email, code); err != nil {
				log.Printf("Failed to send OTP to %s: %v", email, err)
			}
		}()
	} else {
		log.Printf("OTP for %s: %s (email notifier not configured)", email, code)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification code sent to your email"})
}

func (h *AuthHandler) VerifyCode(c *gin.Context) {
	var input VerifyCodeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and 6-digit code required"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))

	otpKey := fmt.Sprintf("stacktrace:otp:%s", email)
	storedCode, err := cache.Client.Get(cache.Ctx, otpKey).Result()
	if err != nil || storedCode != input.Code {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired code"})
		return
	}

	cache.Client.Del(cache.Ctx, otpKey)

	user, err := repository.GetOrCreateUser(h.DB, email)
	if err != nil {
		log.Printf("Failed to get/create user %s: %v", email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	secret := []byte(os.Getenv("JWT_SECRET"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
		},
	})
}
