package controllers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
)

func handleWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Error reading request body: %v\n", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Error reading request body"})
		return
	}

	// This is your Stripe CLI webhook secret for testing your endpoint locally.
	endpointSecret := "whsec_9cbaa6b7a015fc8b3a5ed2d3beae57c67bca68d18b4a90662e3cd3c6d7f625ea"
	event, err := webhook.ConstructEvent(payload, c.GetHeader("Stripe-Signature"), endpointSecret)

	if err != nil {
		log.Printf("Error verifying webhook signature: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error verifying webhook signature"})
		return
	}

	// Handle the event
	switch event.Type {
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
