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
