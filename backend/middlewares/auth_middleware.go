package middlewares

import (
	"errors"
	"log"
	"net/http"
	"organization-management-app/config"
	"organization-management-app/models"
	"organization-management-app/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			err := errors.New("missing Authorization Header")
			log.Println(err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization Header"})
			c.Abort()
			return
		}
		tokenClaim := utils.Claims{}
		token, err := utils.ParseToken(tokenString, &tokenClaim)
		if err != nil {
			err := errors.New("invalid ParseToken")
			log.Println(err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
			c.Abort()
			return
		}
		if !token.Valid {
			err := errors.New("invalid token")
			log.Println(err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Token"})
			c.Abort()
			return
		}

		c.Set("userID", tokenClaim.UserID)
		c.Set("userEmail", tokenClaim.Email)
		c.Set("user", &tokenClaim)
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
		orgID := c.Query("orgId")
		if orgID == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden orgID"})
			c.Abort()
			return
		}
		orgId, err := strconv.ParseUint(orgID, 10, 64)
		if err != nil {
			log.Println("error parse id from string", err)
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden orgID"})
			c.Abort()
			return
		}
		userOrgs := user.(*utils.Claims).Organizations
		role := "Member"
		for _, v := range userOrgs {
			if orgId == uint64(v.Organization.ID) {
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
