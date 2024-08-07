package routes

import (
	"organization-management-app/controllers"
	"organization-management-app/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterOrganizationRoutes(r *gin.RouterGroup) {
	userOrgRoutes := r.Group("/organization", middlewares.AuthMiddleware())
	{
		userOrgRoutes.POST("", controllers.CreateOrganization)
	}
	organizationRoutes := r.Group("/organizations")
	{
		organizationRoutes.Use(middlewares.AuthMiddleware(), middlewares.AdminOnly())
		organizationRoutes.GET("/users", controllers.GetOrganizationsUsers)
		organizationRoutes.GET("/pending", controllers.GetOrganizationsUsersPending)
		organizationRoutes.GET("/subscription-info", controllers.GetOrganizationSubscriptionInfo)
	}
}

func RegisterUserRoutes(r *gin.RouterGroup) {
	userAuthRoutes := r.Group("/users")
	{
		userAuthRoutes.Use(middlewares.AuthMiddleware())
		userAuthRoutes.POST("", middlewares.AdminOnly(), controllers.CreateUser)
		userAuthRoutes.GET("", middlewares.AdminOnly(), controllers.GetUsers)
		userAuthRoutes.POST("/user/free", controllers.CreateUserFreeSubscription)
		userAuthRoutes.POST("/user/subscription", controllers.CreateUserWithSubscription)
		userAuthRoutes.POST("/user/upgrade", controllers.Upgrade)
		userAuthRoutes.POST("/user/downgrade", controllers.Downgrade)
		userAuthRoutes.GET("/profile", controllers.GetProfile)

	}
	userRoutes := r.Group("/user")
	{
		userRoutes.POST("/signup-magic-link", controllers.SignUpWithMagicLink)
		userRoutes.POST("/complete-signup", controllers.CompleteSignup)
		userRoutes.POST("/login-magic-link", controllers.LoginWithMagicLink)
		userRoutes.POST("/login", controllers.MagicLinkLogin)
	}

}
func RegisterProductRoutes(r *gin.RouterGroup) {
	r.POST("/products", controllers.CreateProduct)
	r.GET("/products", controllers.ListProductsWithPrices)
}
func RegisterSubscriptionRoutes(r *gin.RouterGroup) {

	subscriptionRoutes := r.Group("/subscriptions")
	{
		subscriptionRoutes.Use(middlewares.AuthMiddleware(), middlewares.AdminOnly())
		subscriptionRoutes.POST("", controllers.CreateSubscription)
		subscriptionRoutes.GET("", controllers.GetSubscriptions)
		subscriptionRoutes.POST("/prorated-cost", controllers.GetProratedCost)
		subscriptionRoutes.POST("/send-invite", controllers.SendInvite)
		subscriptionRoutes.POST("/disable-user", controllers.DisableUser)
	}

}
func RegisterAuthRoutes(r *gin.RouterGroup) {
	r.POST("subscription/accept-invite", controllers.AcceptInvite)

	r.POST("/invite", controllers.InviteUser)
	r.GET("/verify-magic-link", controllers.VerifyMagicLink)
	r.POST("/singup", controllers.InviteUser)
}
func RegisterWebhookRoutes(r *gin.RouterGroup) {
	r.POST("/webhook", controllers.HandleWebhook)
}
