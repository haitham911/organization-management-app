package middlewares

import (
	"log"
	"net/http"
	"organization-management-app/config"
	"organization-management-app/models"
	"organization-management-app/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization Header"})
			c.Abort()
			return
		}

		token, err := utils.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(utils.Claims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
			c.Abort()
			return
		}
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("user", &user)
		c.Next()
	}
}
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		orgID, ok := c.Get("orgRequest")
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden orgID"})
			c.Abort()
			return
		}
		userOrgs := user.(*utils.Claims).Organizations
		role := ""
		for _, v := range userOrgs {
			if orgID.(uint) == v.Organization.ID {
				role = v.Role
				break
			}
		}
		if role != "Admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// auth Middleware for users check if belong to organizations has active subtractions or users has personal active subtractions
// handel free subtractions request amount
func UserAuthMiddlewareWithFree() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Check if the user belongs to any organization with an active subscription
		hasActiveSubscription := false
		for _, org := range user.(*models.User).Organizations {
			for _, sub := range org.Subscriptions {
				if *sub.Active {
					hasActiveSubscription = true
					break
				}
			}
			if hasActiveSubscription {
				break
			}
		}

		// Check if the user has a personal active subscription
		if !hasActiveSubscription {
			for _, sub := range user.(*models.User).Subscriptions {
				if *sub.Active {
					hasActiveSubscription = true
					break
				}
			}
		}
		// Check if the user has a personal active subscription
		if !hasActiveSubscription {
			userId := user.(*models.User).ID
			usage := user.(*models.User).Usage
			for _, sub := range user.(*models.User).Subscriptions {
				if sub.StripeSubscriptionID == "free" {
					if usage > sub.UsageLimit {
						hasActiveSubscription = false
						break
					}
					if err := config.DB.Model(&models.User{}).Where("id = ?", userId).Update("usage", usage+1).Error; err != nil {
						log.Println(err, "Failed to update user with subscription usage")
						c.JSON(http.StatusForbidden, gin.H{"error": "User does not have an active subscription"})
						c.Abort()
						return
					}
					hasActiveSubscription = true
					break
				}
			}
		}
		if !hasActiveSubscription {
			c.JSON(http.StatusForbidden, gin.H{"error": "User does not have an active subscription"})
			c.Abort()
			return
		}

		c.Next()
	}

}

func UserAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Check if the user belongs to any organization with an active subscription
		hasActiveSubscription := false
		for _, org := range user.(*models.User).Organizations {
			for _, sub := range org.Subscriptions {
				if *sub.Active {
					hasActiveSubscription = true
					break
				}
			}
			if hasActiveSubscription {
				break
			}
		}

		// Check if the user has a personal active subscription
		if !hasActiveSubscription {
			for _, sub := range user.(*models.User).Subscriptions {
				if *sub.Active {
					hasActiveSubscription = true
					break
				}
			}
		}

		if !hasActiveSubscription {
			c.JSON(http.StatusForbidden, gin.H{"error": "User does not have an active subscription"})
			c.Abort()
			return
		}

		c.Next()
	}
}
