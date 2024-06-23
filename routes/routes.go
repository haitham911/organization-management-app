package routes

import (
	"organization-management-app/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterOrganizationRoutes(r *gin.Engine) {
	r.POST("/organizations", controllers.CreateOrganization)
	r.GET("/organizations", controllers.GetOrganizations)
}
func RegisterUserRoutes(r *gin.Engine) {
	r.POST("/users", controllers.CreateUser)
	r.GET("/users", controllers.GetUsers)
}
func RegisterProductRoutes(r *gin.Engine) {
	r.POST("/products", controllers.CreateProduct)
	r.GET("/products", controllers.GetProducts)
}
func RegisterSubscriptionRoutes(r *gin.Engine) {
	r.POST("/subscriptions", controllers.CreateSubscription)
	r.GET("/subscriptions", controllers.GetSubscriptions)
}
func RegisterAuthRoutes(r *gin.Engine) {
	r.POST("/invite", controllers.InviteUser)
	r.GET("/verify-magic-link", controllers.VerifyMagicLink)
}
