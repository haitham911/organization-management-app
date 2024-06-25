package main

import (
	"log"
	"net/http"
	"organization-management-app/config"
	"organization-management-app/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}

	config.InitDB()
	config.InitStripe()

	r := gin.Default()
	// Register API routes
	r.Use(CORSMiddleware()) // this middleware should be applied first so that the auth middleware don't block requests
	v1 := r.Group("/api")
	routes.RegisterOrganizationRoutes(v1)
	routes.RegisterUserRoutes(v1)
	routes.RegisterProductRoutes(v1)
	routes.RegisterSubscriptionRoutes(v1)
	routes.RegisterAuthRoutes(v1)

	if err := r.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Request-Starttime")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
