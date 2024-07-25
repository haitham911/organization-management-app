package controllers

import (
	"fmt"
	"log"
	"net/http"
	"organization-management-app/config"
	"organization-management-app/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/sub"
)

type subscriptionRequest struct {
	UserID          *uint  `json:"user_id"`
	OrganizationID  *uint  `json:"organization_id" binding:"required"`
	PriceID         string `json:"price_id" binding:"required" example:"price_1PVoH4Lq8P7MVUmbz4NnEDsW"`
	Quantity        int    `json:"quantity" binding:"required" example:"1"`
	PaymentMethodID string `json:"payment_method_id" binding:"required"`
	ProductID       uint   `json:"product_id" binding:"required"`
}

// CreateSubscription godoc
// @Summary Create subscription
// @Description  Create subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Security Bearer
// @Param orgId query int  true "Organization Id"
// @Param subscription body subscriptionRequest true "subscriptionRequest"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /subscriptions [post]
func CreateSubscription(c *gin.Context) {
	var subscriptionRequest subscriptionRequest
	if err := c.ShouldBindJSON(&subscriptionRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retrieve the organization from the database
	var organization models.Organization
	if err := config.DB.Where("id = ?", subscriptionRequest.OrganizationID).First(&organization).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	// Attach the payment method to the customer
	_, err := paymentmethod.Attach(subscriptionRequest.PaymentMethodID, &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(organization.StripeCustomerID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to attach payment method"})
		return
	}

	// Set the default payment method for the customer
	_, err = customer.Update(organization.StripeCustomerID, &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(subscriptionRequest.PaymentMethodID),
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set default payment method"})
		return
	}

	// Create the Stripe subscription
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(organization.StripeCustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price:    stripe.String(subscriptionRequest.PriceID),
				Quantity: stripe.Int64(int64(subscriptionRequest.Quantity)),
			},
		},
	}
	subscription, err := sub.New(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	active := subscription.Status == "active"
	// Save the subscription to the database
	dbSubscription := models.Subscription{
		OrganizationID:       subscriptionRequest.OrganizationID,
		StripeSubscriptionID: subscription.ID,
		SubscriptionStatus:   string(subscription.Status),
		Quantity:             subscriptionRequest.Quantity,
		Active:               &active,
		PriceId:              subscriptionRequest.PriceID,
		ProductID:            subscriptionRequest.ProductID,
	}
	if err := config.DB.Create(&dbSubscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscriptionId": subscription.ID})
}

// GetSubscriptions godoc
// @Summary get subscription
// @Description  Create subscription
// @Tags subscriptions
// @Accept json
// @Security Bearer
// @Produce json
// @Param orgId query int  true "Organization Id"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /subscriptions [get]
func GetSubscriptions(c *gin.Context) {
	orgID := c.Query("orgId")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orgId required"})
		return
	}
	orgId, err := strconv.ParseUint(orgID, 10, 64)
	if err != nil {
		log.Println("error parse id from string", err)
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden orgID"})
		c.Abort()
		return
	}
	var subscriptions []models.Subscription
	config.DB.Where("organization_id  =?", orgId).Find(&subscriptions)
	c.JSON(http.StatusOK, subscriptions)
}

// ActivateSubscription activates a subscription in the database.
func ActivateSubscription(stripeSubscriptionID string) error {
	// Find the subscription by StripeSubscriptionID
	var subscription models.Subscription
	if err := config.DB.Where("stripe_subscription_id = ?", stripeSubscriptionID).First(&subscription).Error; err != nil {
		return fmt.Errorf("could not find subscription: %w", err)
	}

	// Update the subscription to active
	*subscription.Active = true
	if err := config.DB.Save(&subscription).Error; err != nil {
		return fmt.Errorf("could not activate subscription: %w", err)
	}

	return nil
}
