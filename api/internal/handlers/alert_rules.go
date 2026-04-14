package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"stacktrace/internal/models"
	"stacktrace/internal/repository"
)

type AlertRuleHandler struct {
	DB *sql.DB
}

func NewAlertRuleHandler(db *sql.DB) *AlertRuleHandler {
	return &AlertRuleHandler{DB: db}
}

type CreateAlertRuleInput struct {
	Condition     string `json:"condition" binding:"required"`
	Threshold     int    `json:"threshold" binding:"required,min=1"`
	WindowSeconds int    `json:"window_seconds" binding:"required,min=10,max=3600"`
	Channel       string `json:"channel" binding:"required,oneof=email whatsapp"`
	Destination   string `json:"destination" binding:"required"`
}

func (h *AlertRuleHandler) Create(c *gin.Context) {
	projectID, ok := getProjectIDFromContext(c)
	if !ok {
		return
	}

	var input CreateAlertRuleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule := &models.AlertRule{
		ProjectID:     projectID,
		Condition:     input.Condition,
		Threshold:     input.Threshold,
		WindowSeconds: input.WindowSeconds,
		Channel:       input.Channel,
		Destination:   input.Destination,
	}

	if err := repository.CreateAlertRule(h.DB, rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create alert rule"})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

func (h *AlertRuleHandler) List(c *gin.Context) {
	projectID, ok := getProjectIDFromContext(c)
	if !ok {
		return
	}

	rules, err := repository.ListAlertRules(h.DB, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list alert rules"})
		return
	}

	if rules == nil {
		rules = []models.AlertRule{}
	}

	c.JSON(http.StatusOK, gin.H{"data": rules})
}

func (h *AlertRuleHandler) Delete(c *gin.Context) {
	projectID, ok := getProjectIDFromContext(c)
	if !ok {
		return
	}

	ruleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}

	if err := repository.DeleteAlertRule(h.DB, ruleID, projectID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert rule deleted"})
}
