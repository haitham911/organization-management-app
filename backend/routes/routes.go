package routes

import (
	"organization-management-app/controllers"
	"organization-management-app/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterOrganizationRoutes(r *gin.RouterGroup) {
	r.POST("/organizations", controllers.CreateOrganization)
	r.GET("/organizations", controllers.GetOrganizations)
}
func RegisterUserRoutes(r *gin.RouterGroup) {
	r.Use(middlewares.AuthMiddleware())
	r.POST("/users", middlewares.AdminOnly(), controllers.CreateUser)
	r.GET("/users", middlewares.AdminOnly(), controllers.GetUsers)
	r.POST("/user/free", controllers.CreateUserFreeSubscription)
	r.POST("/user/subscription", controllers.CreateUserWithSubscription)
	r.POST("/user/upgrade", controllers.Upgrade)
	r.POST("/user/downgrade", controllers.Downgrade)

}
func RegisterProductRoutes(r *gin.RouterGroup) {
	r.POST("/products", controllers.CreateProduct)
	r.GET("/products", controllers.ListProductsWithPrices)
}
func RegisterSubscriptionRoutes(r *gin.RouterGroup) {
	r.POST("/subscriptions", controllers.CreateSubscription)
	r.GET("/subscriptions", controllers.GetSubscriptions)
}
func RegisterAuthRoutes(r *gin.RouterGroup) {
	r.POST("/invite", controllers.InviteUser)
	r.GET("/verify-magic-link", controllers.VerifyMagicLink)
}
func RegisterWebhookRoutes(r *gin.RouterGroup) {
	r.POST("/webhook", controllers.HandleWebhook)
}
