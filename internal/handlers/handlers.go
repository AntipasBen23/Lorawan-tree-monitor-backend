package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/AntipasBen23/lorawan-tree-monitor/config"
	"github.com/AntipasBen23/lorawan-tree-monitor/internal/database"
	"github.com/AntipasBen23/lorawan-tree-monitor/internal/models"
	"github.com/AntipasBen23/lorawan-tree-monitor/internal/token"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	db              *database.DB
	tokenCalculator *token.Calculator
	config          *config.Config
}

func NewHandler(db *database.DB, tokenCalc *token.Calculator, cfg *config.Config) *Handler {
	return &Handler{
		db:              db,
		tokenCalculator: tokenCalc,
		config:          cfg,
	}
}

// LoRaWANWebhook handles incoming data from the LoRaWAN Network Server
func (h *Handler) LoRaWANWebhook(c *gin.Context) {
	var payload models.LoRaWANPayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("Invalid payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	log.Printf("Received data from device: %s", payload.DeviceID)

	// Get tree by sensor ID
	tree, err := h.db.GetTreeBySensorID(payload.DeviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Unknown device: %s", payload.DeviceID)
			c.JSON(http.StatusNotFound, gin.H{"error": "Device not registered"})
			return
		}
		log.Printf("Database error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Parse measurement data from payload
	measurement := &models.Measurement{
		TreeID:       tree.ID,
		SoilMoisture: getFloatFromPayload(payload.Data, "soil_moisture"),
		Temperature:  getFloatFromPayload(payload.Data, "temperature"),
		Tilt:         getFloatFromPayload(payload.Data, "tilt"),
		BatteryLevel: getFloatFromPayload(payload.Data, "battery_level"),
	}

	// Parse timestamp
	parsedTime, err := time.Parse(time.RFC3339, payload.Timestamp)
	if err != nil {
		measurement.Timestamp = time.Now()
	} else {
		measurement.Timestamp = parsedTime
	}

	// Validate measurement
	if !h.tokenCalculator.ValidateMeasurement(measurement) {
		log.Printf("Invalid measurement data from device %s", payload.DeviceID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid measurement data"})
		return
	}

	// Check for duplicate
	isDuplicate, err := h.db.CheckDuplicateMeasurement(tree.ID, measurement.Timestamp)
	if err != nil {
		log.Printf("Error checking duplicate: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if isDuplicate {
		log.Printf("Duplicate measurement detected for tree %d at %v", tree.ID, measurement.Timestamp)
		c.JSON(http.StatusOK, gin.H{"message": "Duplicate measurement ignored"})
		return
	}

	// Calculate tokens
	tokensAwarded := h.tokenCalculator.CalculateTokens(measurement)
	measurement.TokensAwarded = tokensAwarded

	// Store measurement
	if err := h.db.InsertMeasurement(measurement); err != nil {
		log.Printf("Error storing measurement: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store measurement"})
		return
	}

	// Update user token balance
	if err := h.db.UpdateUserTokenBalance(tree.UserID, tokensAwarded); err != nil {
		log.Printf("Error updating token balance: %v", err)
	}

	// Record token transaction
	transaction := &models.TokenTransaction{
		UserID:        tree.UserID,
		MeasurementID: measurement.ID,
		Amount:        tokensAwarded,
		Type:          "earn",
	}

	if err := h.db.InsertTokenTransaction(transaction); err != nil {
		log.Printf("Error recording transaction: %v", err)
	}

	log.Printf("Measurement stored. Awarded %d tokens to user %d", tokensAwarded, tree.UserID)

	c.JSON(http.StatusOK, gin.H{
		"message":        "Measurement received",
		"tokens_awarded": tokensAwarded,
	})
}

// GetTrees returns all trees
func (h *Handler) GetTrees(c *gin.Context) {
	trees, err := h.db.GetAllTrees()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trees"})
		return
	}

	c.JSON(http.StatusOK, trees)
}

// GetTreeMeasurements returns measurement history for a tree
func (h *Handler) GetTreeMeasurements(c *gin.Context) {
	treeIDStr := c.Param("id")
	treeID, err := strconv.Atoi(treeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tree ID"})
		return
	}

	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	measurements, err := h.db.GetMeasurementsByTreeID(treeID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch measurements"})
		return
	}

	c.JSON(http.StatusOK, measurements)
}

// GetUserTokenBalance returns a user's token balance
func (h *Handler) GetUserTokenBalance(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": user.ID,
		"username": user.Username,
		"token_balance": user.TokenBalance,
	})
}

// HealthCheck returns server health status
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"timestamp": time.Now(),
	})
}

// Helper function to extract float values from payload data
func getFloatFromPayload(data map[string]interface{}, key string) float64 {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			var f float64
			json.Unmarshal([]byte(v), &f)
			return f
		}
	}
	return 0.0
}