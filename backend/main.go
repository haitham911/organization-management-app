package main

import (
	"log"
	"net/http"
	"organization-management-app/config"
	_ "organization-management-app/docs"
	"organization-management-app/routes"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
)

// godoc
//
//	@title			organization management
//	@version		1.0
//	@description	organization-management-app
//	@contact.name	Haitham
//	@contact.url	https://github.com/haitham911/organization-management-app
//	@BasePath		/api/v1
//	@securityDefinitions.apikey Bearer
//	@in header
//	@name Authorization

// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}
	// Verify the JWT_SECRET environment variable is set
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatalf("JWT_SECRET environment variable not set")
	}
	config.InitDB()
	config.InitStripe()

	r := gin.Default()
	ginSwaggerConfig := &ginSwagger.Config{
		URL:                  "/swagger/doc.json",
		PersistAuthorization: true,
	}
	swaggerUIOpts := ginSwagger.CustomWrapHandler(ginSwaggerConfig, swaggerFiles.Handler)
	r.GET("/swagger/*any", swaggerUIOpts)
	// Register API routes
	r.Use(CORSMiddleware()) // this middleware should be applied first so that the auth middleware don't block requests
	v1 := r.Group("/api/v1")
	routes.RegisterOrganizationRoutes(v1)
	routes.RegisterUserRoutes(v1)
	routes.RegisterProductRoutes(v1)
	routes.RegisterSubscriptionRoutes(v1)
	routes.RegisterAuthRoutes(v1)
	routes.RegisterWebhookRoutes(v1)
	if err := r.Run(":" + os.Getenv("PORT")); err != nil {
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
