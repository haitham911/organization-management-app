package controllers

import (
	"net/http"
	"organization-management-app/config"
	"organization-management-app/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(c *gin.Context) {
	var userRequest struct {
		Name                 string `json:"name" binding:"required"`
		Email                string `json:"email" binding:"required"`
		Password             string `json:"password" binding:"required"`
		OrganizationID       uint   `json:"organization_id" binding:"required"`
		StripeSubscriptionID string `json:"stripe_subscription_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the organization can add more subscriptions
	var subscription models.Subscription
	if err := config.DB.Where("organization_id = ? AND stripe_subscription_id = ?", userRequest.OrganizationID, userRequest.StripeSubscriptionID).First(&subscription).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	var totalUsers int64
	if err := config.DB.Model(&models.UserOrganization{}).Where("organization_id = ? AND stripe_subscription_id = ?", userRequest.OrganizationID, userRequest.StripeSubscriptionID).Count(&totalUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count users"})
		return
	}

	if totalUsers >= int64(subscription.Quantity) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Organization cannot add more subscriptions"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Name:     userRequest.Name,
		Email:    userRequest.Email,
		Password: string(hashedPassword),
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Associate user with the organization and their Stripe subscription
	userOrg := models.UserOrganization{
		UserID:               user.ID,
		OrganizationID:       userRequest.OrganizationID,
		StripeSubscriptionID: userRequest.StripeSubscriptionID,
	}
	if err := config.DB.Create(&userOrg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate user with organization"})
		return
	}

	c.JSON(http.StatusOK, user)
}
func GetUsers(c *gin.Context) {
	var users []models.User
	config.DB.Preload("Organizations").Find(&users)
	c.JSON(http.StatusOK, users)
}

// Check if users belong to an organization have subscription to a product
func UsersHaveSubscriptionToProduct(c *gin.Context) {
	var request struct {
		OrganizationID uint `json:"organization_id" binding:"required"`
		ProductID      uint `json:"product_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var organization models.Organization
	if err := config.DB.Where("id = ?", request.OrganizationID).First(&organization).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	var subscriptions []models.Subscription
	if err := config.DB.Where("organization_id = ? AND product_id = ?", request.OrganizationID, request.ProductID).Find(&subscriptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find subscriptions"})
		return
	}

	hasSubscription := len(subscriptions) > 0

	c.JSON(http.StatusOK, gin.H{"has_subscription": hasSubscription})
}
