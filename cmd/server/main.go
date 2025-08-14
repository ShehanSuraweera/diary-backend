package main

import (
	"diary-backend/internal/config"
	"diary-backend/internal/database"
	"diary-backend/internal/routes"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Connect to database
	if err := database.Connect(&cfg.Database); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router, cfg.CORS.AllowedOrigins)

	// Start server
	serverAddr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Starting server on %s", serverAddr)

	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}
}
