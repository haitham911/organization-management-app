package controllers

import (
	"net/http"
	"organization-management-app/config"
	"organization-management-app/models"

	"github.com/gin-gonic/gin"
)

func CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	config.DB.Create(&user)
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
