package main

import (
	"encoding/json"
	"log"
	"nucleus-api/handlers"
	"nucleus-api/middleware"
	"nucleus-api/models"
	redispkg "nucleus-api/redis"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("[NUCLEUS-API] Starting API server...")

	// Redis connection
	redisURL := os.Getenv("NUCLEUS_REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://redis:6379"
	}

	subscriber := redispkg.NewSubscriber(redisURL)
	defer subscriber.Close()

	// Subscribe to nucleus channels
	subscriber.Subscribe("nucleus:executions", "nucleus:rollbacks", "nucleus:sessions")

	// Set up handlers
	handlers.RedisSubscriber = subscriber
	handlers.RollbackBroadcast = subscriber.Publish

	// Handle incoming execution events from Redis
	go func() {
		ch := subscriber.AddListener("api-server")
		for msg := range ch {
			switch msg.Type {
			case "execution_complete":
				data, err := json.Marshal(msg.Data)
				if err != nil {
					continue
				}
				var exec models.ExecutionNode
				if err := json.Unmarshal(data, &exec); err != nil {
					continue
				}
				handlers.AddExecution(exec)
			case "session_started":
				data, err := json.Marshal(msg.Data)
				if err != nil {
					continue
				}
				var session models.Session
				if err := json.Unmarshal(data, &session); err != nil {
					continue
				}
				handlers.AddSession(session)
			}
		}
	}()

	// Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.LoggerMiddleware())

	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-Nucleus-Key, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health + metrics (no auth)
	r.GET("/health", handlers.GetHealthReady)
	r.GET("/health/ready", handlers.GetHealthReady)
	r.GET("/metrics", handlers.GetMetrics)

	// API v1 routes
	v1 := r.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware())
	{
		// Executions
		v1.GET("/executions", handlers.GetExecutions)
		v1.GET("/executions/:id", handlers.GetExecution)
		v1.GET("/executions/:id/context", handlers.GetExecutionContext)

		// Context
		v1.GET("/context", handlers.GetContext)
		v1.GET("/context/graph", handlers.GetContextGraph)

		// Rollback
		v1.POST("/rollback/:execution_id", handlers.PostRollback)

		// Sessions
		v1.GET("/sessions", handlers.GetSessions)
		v1.POST("/sessions", handlers.CreateSession)
		v1.GET("/sessions/:id/replay", handlers.GetSessionReplay)

		// Agent
		v1.POST("/agent/execute", handlers.AgentExecute)
		v1.POST("/agent/plan", handlers.AgentPlan)

		// Skills
		v1.GET("/skills", handlers.GetSkills)
		v1.GET("/skills/:name", handlers.GetSkill)
		v1.POST("/skills/:name/run", handlers.RunSkill)

		// Context environment
		v1.GET("/context/environment", handlers.GetContext)

		// Rollback preview
		v1.GET("/rollback/:execution_id/preview", handlers.PostRollback)

		// Session detail and export
		v1.GET("/sessions/:id", handlers.GetSessionReplay)
		v1.GET("/sessions/:id/export", handlers.GetSessionReplay)

		// Natural language
		v1.POST("/natural/translate", handlers.NaturalTranslate)

		// Search
		v1.GET("/search", handlers.SearchExecutions)

		// WebSocket
		v1.GET("/ws/stream", handlers.WSStream)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("[NUCLEUS-API] Listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
