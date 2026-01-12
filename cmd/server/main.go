package main

import (
	"log"
	"os"

	"github.com/AntipasBen23/lorawan-tree-monitor/config"
	"github.com/AntipasBen23/lorawan-tree-monitor/internal/database"
	"github.com/AntipasBen23/lorawan-tree-monitor/internal/handlers"
	"github.com/AntipasBen23/lorawan-tree-monitor/internal/token"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("==> Starting application...")

	// Load configuration
	cfg := config.Load()
	log.Printf("==> Configuration loaded - Port: %s", cfg.ServerPort)
	log.Printf("==> Database URL (first 30 chars): %.30s...", cfg.DatabaseURL)

	// Connect to database
	log.Println("==> Attempting database connection...")
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Printf("ERROR: Failed to connect to database: %v", err)
		log.Printf("ERROR: Database URL length: %d", len(cfg.DatabaseURL))
		os.Exit(1)
	}
	defer db.Close()
	log.Println("==> Database connected successfully!")

	// Initialize database schema
	log.Println("==> Initializing database schema...")
	if err := db.InitSchema(); err != nil {
		log.Printf("ERROR: Failed to initialize schema: %v", err)
		os.Exit(1)
	}
	log.Println("==> Database schema initialized!")

	// Initialize token calculator
	tokenCalc := token.NewCalculator(cfg.TokensPerReading)
	log.Printf("==> Token calculator initialized (tokens per reading: %d)", cfg.TokensPerReading)

	// Initialize handlers
	handler := handlers.NewHandler(db, tokenCalc, cfg)
	log.Println("==> Handlers initialized!")

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

	log.Println("==> Routes configured!")

	// Start server
	addr := "0.0.0.0:" + cfg.ServerPort
	log.Printf("==> Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("ERROR: Failed to start server: %v", err)
	}
}