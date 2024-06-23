package controllers

import (
	"net/http"
	"organization-management-app/config"
	"organization-management-app/models"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/sub"
)

func CreateSubscription(c *gin.Context) {
	var subscriptionRequest struct {
		OrganizationID uint `json:"organization_id"`
		ProductID      uint `json:"product_id"`
		Quantity       int  `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&subscriptionRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var organization models.Organization
	var product models.Product
	if err := config.DB.First(&organization, subscriptionRequest.OrganizationID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}
	if err := config.DB.First(&product, subscriptionRequest.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	params := &stripe.SubscriptionParams{
		Customer: stripe.String(organization.StripeCustomerID), // Replace with actual Stripe customer ID
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price:    stripe.String(product.PriceID),
				Quantity: stripe.Int64(int64(subscriptionRequest.Quantity)),
			},
		},
	}
	s, err := sub.New(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	subscription := models.Subscription{
		OrganizationID:       subscriptionRequest.OrganizationID,
		ProductID:            subscriptionRequest.ProductID,
		StripeSubscriptionID: s.ID,
		Quantity:             subscriptionRequest.Quantity,
	}
	config.DB.Create(&subscription)
	c.JSON(http.StatusOK, subscription)
}

func GetSubscriptions(c *gin.Context) {
	var subscriptions []models.Subscription
	config.DB.Find(&subscriptions)
	c.JSON(http.StatusOK, subscriptions)
}
