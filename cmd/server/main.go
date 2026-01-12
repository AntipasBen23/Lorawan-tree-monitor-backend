package main

import (
	"log"

	"github.com/AntipasBen23/lorawan-tree-monitor/config"
	"github.com/AntipasBen23/lorawan-tree-monitor/internal/database"
	"github.com/AntipasBen23/lorawan-tree-monitor/internal/handlers"
	"github.com/AntipasBen23/lorawan-tree-monitor/internal/token"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()
	log.Println("Configuration loaded")

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize database schema
	if err := db.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Initialize token calculator
	tokenCalc := token.NewCalculator(cfg.TokensPerReading)

	// Initialize handlers
	handler := handlers.NewHandler(db, tokenCalc, cfg)

	// Setup router
	router := gin.Default()

	// Routes
	router.GET("/health", handler.HealthCheck)
	router.POST("/webhook/lorawan", handler.LoRaWANWebhook)
	
	api := router.Group("/api/v1")
	{
		api.GET("/trees", handler.GetTrees)
		api.GET("/trees/:id/measurements", handler.GetTreeMeasurements)
		api.GET("/users/:id/tokens", handler.GetUserTokenBalance)
	}

	// Start server
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}