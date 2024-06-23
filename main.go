package main

import (
	"log"
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

	// Serve static files from the "frontend/build" directory
	r.Static("/static", "./frontend/build/static")
	r.StaticFile("/", "./frontend/build/index.html")
	r.StaticFile("/favicon.ico", "./frontend/build/favicon.ico")
	r.StaticFile("/manifest.json", "./frontend/build/manifest.json")

	// Register API routes
	routes.RegisterOrganizationRoutes(r)
	routes.RegisterUserRoutes(r)
	routes.RegisterProductRoutes(r)
	routes.RegisterSubscriptionRoutes(r)
	routes.RegisterAuthRoutes(r)

	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/build/index.html")
	})

	if err := r.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
