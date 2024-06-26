package controllers

import (
	"log"
	"net/http"
	"organization-management-app/config"
	"organization-management-app/models"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
)

func CreateOrganization(c *gin.Context) {
	var organization models.Organization
	if err := c.ShouldBindJSON(&organization); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Create Stripe Customer
	params := &stripe.CustomerParams{
		Name:  stripe.String(organization.Name),
		Email: stripe.String(organization.Email), // Assuming the organization has an email field
	}
	stripeCustomer, err := customer.New(params)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Stripe customer"})
		return
	}

	organization.StripeCustomerID = stripeCustomer.ID
	config.DB.Create(&organization)
	c.JSON(http.StatusOK, organization)
}
func GetOrganizations(c *gin.Context) {
	var organizations []models.Organization
	config.DB.Preload("Users").Find(&organizations)
	c.JSON(http.StatusOK, organizations)
}

// Check if an organization can add more users
func CanAddMoreUsers(c *gin.Context) {
	var request struct {
		OrganizationID uint `json:"organization_id" binding:"required"`
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

	var totalUsers int64
	if err := config.DB.Model(&models.User{}).Where("organization_id = ?", request.OrganizationID).Count(&totalUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count users"})
		return
	}

	var totalSubscriptions int64
	if err := config.DB.Model(&models.Subscription{}).Where("organization_id = ?", request.OrganizationID).First("quantity", &totalSubscriptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count subscriptions"})
		return
	}

	canAddMoreUsers := totalUsers < totalSubscriptions

	c.JSON(http.StatusOK, gin.H{"can_add_more_users": canAddMoreUsers})
}
