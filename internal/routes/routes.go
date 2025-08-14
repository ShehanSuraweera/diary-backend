package routes

import (
	"diary-backend/internal/handlers"
	"diary-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, corsOrigins string) {
	// Add CORS middleware
	router.Use(middleware.CORS(corsOrigins))

	// API version 1
	v1 := router.Group("/api/v1")
	{
		// Task routes
		tasks := v1.Group("/tasks")
		{
			tasks.GET("", handlers.GetTasks)           // GET /api/v1/tasks
			tasks.GET("/:id", handlers.GetTaskByID)    // GET /api/v1/tasks/:id
			tasks.POST("", handlers.CreateTask)        // POST /api/v1/tasks
			tasks.PUT("/:id", handlers.UpdateTask)     // PUT /api/v1/tasks/:id
			tasks.DELETE("/:id", handlers.DeleteTask)  // DELETE /api/v1/tasks/:id
			tasks.GET("/stats", handlers.GetTaskStats) // GET /api/v1/tasks/stats
		}
		// Learning Resources routes
		resources := v1.Group("/resources")
		{
			resources.GET("", handlers.GetResources)                      // GET /api/v1/resources
			resources.GET("/:id", handlers.GetResourceByID)               // GET /api/v1/resources/:id
			resources.POST("", handlers.CreateResource)                   // POST /api/v1/resources
			resources.PUT("/:id", handlers.UpdateResource)                // PUT /api/v1/resources/:id
			resources.DELETE("/:id", handlers.DeleteResource)             // DELETE /api/v1/resources/:id
			resources.PATCH("/:id/status", handlers.UpdateResourceStatus) // PATCH /api/v1/resources/:id/status
			resources.PATCH("/:id/rating", handlers.UpdateResourceRating) // PATCH /api/v1/resources/:id/rating
			resources.GET("/stats", handlers.GetResourceStats)            // GET /api/v1/resources/stats
			resources.GET("/technologies", handlers.GetTechnologies)      // GET /api/v1/resources/technologies
			resources.POST("/import-url", handlers.ImportFromURL)         // POST /api/v1/resources/import-url
		}

	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "diary-backend",
		})
	})
}
