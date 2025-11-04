package main

import (
	"log"
	"os"
	"os/signal"
	"sms-gateway/config"
	"sms-gateway/database"
	"sms-gateway/handlers"
	"sms-gateway/middleware"
	"sms-gateway/utils"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Start usage reset scheduler
	utils.StartUsageResetScheduler()

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Configure appropriately for production
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Key", "X-API-Secret"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initialize handlers
	smsHandler := handlers.NewSMSHandler()
	clientHandler := handlers.NewClientHandler()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "SMS Gateway API",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// SMS endpoints (require API key authentication)
		sms := v1.Group("/sms")
		sms.Use(middleware.APIKeyAuth())
		{
			sms.POST("/send", smsHandler.SendSingleSMS)
			sms.POST("/send/bulk", smsHandler.SendBulkSMS)
			sms.GET("/logs", smsHandler.GetSMSLogs)
			sms.GET("/stats", smsHandler.GetStats)
		}

		// Admin endpoints (require Basic Auth)
		admin := v1.Group("/admin")
		admin.Use(middleware.BasicAuth())
		{
			admin.POST("/clients", clientHandler.CreateClient)
			admin.GET("/clients", clientHandler.ListClients)
			admin.PUT("/clients/:id", clientHandler.UpdateClient)
			admin.POST("/clients/:id/reset", clientHandler.ResetClientUsage)
		}
	}

	// Start server
	addr := config.AppConfig.ServerHost + ":" + config.AppConfig.ServerPort
	log.Printf("Starting SMS Gateway API server on %s", addr)

	// Graceful shutdown
	go func() {
		if err := router.Run(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}

