package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"organization-management-app/config"
	"organization-management-app/models"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
)

func HandleWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Error reading request body: %v\n", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Error reading request body"})
		return
	}
	log.Println("payload", string(payload))
	// This is your Stripe CLI webhook secret for testing your endpoint locally.
	endpointSecret := os.Getenv("STRIPE_ENDPOINT_SECRET")
	event, err := webhook.ConstructEvent(payload, c.GetHeader("Stripe-Signature"), endpointSecret)

	if err != nil {
		log.Printf("Error verifying webhook signature: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error verifying webhook signature"})
		return
	}

	// Handle the event
	switch event.Type {
	case "customer.subscription.created", "customer.subscription.updated", "customer.subscription.deleted":
		var subscription stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
			log.Printf("Error parsing webhook JSON: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing webhook JSON"})
			return
		}

		// Update the subscription status in your database
		if err := updateSubscriptionStatus(subscription); err != nil {
			log.Printf("Error updating subscription status: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating subscription status"})
			return
		}

		log.Printf("Subscription %s was updated to status: %s", subscription.ID, subscription.Status)

	case "invoice.payment_succeeded":
		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			log.Printf("Error parsing webhook JSON: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing webhook JSON"})
			return
		}

		// Activate the subscription for the customer
		err := ActivateSubscription(invoice.Subscription.ID)
		if err != nil {
			log.Printf("Error activating subscription: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error activating subscription"})
			return
		}
		log.Printf("Invoice payment succeeded for customer: %s", invoice.CustomerEmail)
	default:
		log.Printf("Unhandled event type: %s\n", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
func updateSubscriptionStatus(subscription stripe.Subscription) error {
	var dbSubscription models.Subscription
	if err := config.DB.Where("stripe_subscription_id = ?", subscription.ID).First(&dbSubscription).Error; err != nil {
		return fmt.Errorf("could not find subscription: %w", err)
	}

	dbSubscription.Active = subscription.Status == "active"
	dbSubscription.SubscriptionStatus = string(subscription.Status)
	if err := config.DB.Save(&dbSubscription).Error; err != nil {
		return fmt.Errorf("could not update subscription status: %w", err)
	}

	return nil
}
