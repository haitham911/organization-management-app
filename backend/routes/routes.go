package routes

import (
	"organization-management-app/controllers"
	"organization-management-app/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterOrganizationRoutes(r *gin.RouterGroup) {
	organizationRoutes := r.Group("/organizations")
	{
		organizationRoutes.POST("", controllers.CreateOrganization)
		organizationRoutes.GET("", controllers.GetOrganizations)
		organizationRoutes.GET("/subscription-info", controllers.GetOrganizationSubscriptionInfo)
	}

}
func RegisterUserRoutes(r *gin.RouterGroup) {
	userRoutes := r.Group("/users")
	{
		userRoutes.Use(middlewares.AuthMiddleware())
		userRoutes.POST("", middlewares.AdminOnly(), controllers.CreateUser)
		userRoutes.GET("", middlewares.AdminOnly(), controllers.GetUsers)
		userRoutes.POST("/user/free", controllers.CreateUserFreeSubscription)
		userRoutes.POST("/user/subscription", controllers.CreateUserWithSubscription)
		userRoutes.POST("/user/upgrade", controllers.Upgrade)
		userRoutes.POST("/user/downgrade", controllers.Downgrade)

	}

}
func RegisterProductRoutes(r *gin.RouterGroup) {
	r.POST("/products", controllers.CreateProduct)
	r.GET("/products", controllers.ListProductsWithPrices)
}
func RegisterSubscriptionRoutes(r *gin.RouterGroup) {
	subscriptionRoutes := r.Group("/subscriptions")
	{
		subscriptionRoutes.POST("", controllers.CreateSubscription)
		subscriptionRoutes.GET("/subscriptions", controllers.GetSubscriptions)
		subscriptionRoutes.POST("/prorated-cost", controllers.GetProratedCost)
		subscriptionRoutes.POST("/send-invite", controllers.SendInvite)
		subscriptionRoutes.POST("/accept-invite", controllers.AcceptInvite)
		subscriptionRoutes.POST("/disable-user", controllers.DisableUser)
	}

}
func RegisterAuthRoutes(r *gin.RouterGroup) {
	r.POST("/invite", controllers.InviteUser)
	r.GET("/verify-magic-link", controllers.VerifyMagicLink)
}
func RegisterWebhookRoutes(r *gin.RouterGroup) {
	r.POST("/webhook", controllers.HandleWebhook)
}
