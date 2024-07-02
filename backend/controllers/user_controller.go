package controllers

import (
	"log"
	"net/http"
	"organization-management-app/config"
	"organization-management-app/models"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/sub"
	"golang.org/x/crypto/bcrypt"
)

const FreeLimit = 100

func CreateUser(c *gin.Context) {
	var userRequest struct {
		Name                 string `json:"name" binding:"required"`
		Email                string `json:"email" binding:"required"`
		Password             string `json:"password" binding:"required"`
		Role                 string `json:"role" binding:"required"`
		OrganizationID       uint   `json:"organization_id"`
		StripeSubscriptionID string `json:"stripe_subscription_id"`
	}

	if err := c.ShouldBindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if userRequest.Role != "Admin" {
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
		Role:     userRequest.Role,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	if userRequest.Role != "Admin" {
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

// CreateUserWithoutOrganization handles the creation of users with free  subscription
func CreateUserFreeSubscription(c *gin.Context) {
	var userRequest struct {
		Name       string `json:"name" binding:"required"`
		Email      string `json:"email" binding:"required"`
		Password   string `json:"password" binding:"required"`
		Role       string `json:"role" binding:"required"`
		UsageLimit int    `json:"usage_limit" binding:"required"`
		ProductID  uint   `json:"product_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		Role:     userRequest.Role,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	subscription := models.Subscription{
		UserID:               &user.ID,
		ProductID:            userRequest.ProductID,
		UsageLimit:           FreeLimit,
		StripeSubscriptionID: "free",
		Quantity:             1,
		Active:               stripe.Bool(true),
		PriceId:              "free",
		SubscriptionStatus:   "free",
	}

	if err := config.DB.Create(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateUserWithoutOrganization handles the creation of users with their own individual subscription
func CreateUserWithSubscription(c *gin.Context) {
	var userRequest struct {
		Name                 string `json:"name" binding:"required"`
		Email                string `json:"email" binding:"required"`
		Password             string `json:"password" binding:"required"`
		Role                 string `json:"role" binding:"required"`
		UsageLimit           int    `json:"usage_limit" binding:"required"`
		ProductID            uint   `json:"product_id" binding:"required"`
		PriceID              string `json:"price_id" binding:"required"`
		PaymentMethodID      string `json:"payment_method_id" binding:"required"`
		StripeSubscriptionID string `json:"stripe_subscription_id"`
	}

	if err := c.ShouldBindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		Role:     userRequest.Role,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Create Stripe Customer
	params := &stripe.CustomerParams{
		Name:  stripe.String(user.Name),
		Email: stripe.String(user.Email),
	}
	stripeCustomer, err := customer.New(params)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Stripe customer"})
		return
	}

	user.StripeCustomerID = stripeCustomer.ID
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user with subscription"})
		return
	}

	// Attach the payment method to the customer
	_, err = paymentmethod.Attach(userRequest.PaymentMethodID, &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(user.StripeCustomerID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to attach payment method"})
		return
	}

	// Set the default payment method for the customer
	_, err = customer.Update(user.StripeCustomerID, &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(userRequest.PaymentMethodID),
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set default payment method"})
		return
	}

	// Create the Stripe subscription
	subParams := &stripe.SubscriptionParams{
		Customer: stripe.String(user.StripeCustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price:    stripe.String(userRequest.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
	}
	subscription, err := sub.New(subParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	active := subscription.Status == "active"
	// Save the subscription to the database
	dbSubscription := models.Subscription{
		UserID:               &user.ID,
		StripeSubscriptionID: subscription.ID,
		SubscriptionStatus:   string(subscription.Status),
		Quantity:             1,
		Active:               &active,
		PriceId:              userRequest.PriceID,
		ProductID:            userRequest.ProductID,
	}
	if err := config.DB.Create(&dbSubscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save subscription"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// Upgrade handles the  users from free to subscription
func Upgrade(c *gin.Context) {
	var upgradeRequest struct {
		UserID          *uint  `json:"user_id"`
		PriceID         string `json:"price_id" binding:"required"`
		PaymentMethodID string `json:"payment_method_id" binding:"required"`
		ProductID       uint   `json:"product_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&upgradeRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retrieve the user from the database
	var user models.User
	if err := config.DB.Where("id = ?", upgradeRequest.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if user.StripeCustomerID == "" {
		// Create Stripe Customer
		params := &stripe.CustomerParams{
			Name:  stripe.String(user.Name),
			Email: stripe.String(user.Email),
		}
		stripeCustomer, err := customer.New(params)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Stripe customer"})
			return
		}

		user.StripeCustomerID = stripeCustomer.ID
		if err := config.DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user with subscription"})
			return
		}

	}
	// Attach the payment method to the customer
	_, err := paymentmethod.Attach(upgradeRequest.PaymentMethodID, &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(user.StripeCustomerID),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to attach payment method"})
		return
	}

	// Set the default payment method for the customer
	_, err = customer.Update(user.StripeCustomerID, &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(upgradeRequest.PaymentMethodID),
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set default payment method"})
		return
	}

	// Create the Stripe subscription
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(user.StripeCustomerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price:    stripe.String(upgradeRequest.PriceID),
				Quantity: stripe.Int64(1),
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
		UserID:               upgradeRequest.UserID,
		StripeSubscriptionID: subscription.ID,
		SubscriptionStatus:   string(subscription.Status),
		Quantity:             1,
		Active:               &active,
		PriceId:              upgradeRequest.PriceID,
		ProductID:            upgradeRequest.ProductID,
	}
	if err := config.DB.Create(&dbSubscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscriptionId": subscription.ID})
}

// Downgrade handles the  users from subscription to free
func Downgrade(c *gin.Context) {
	var downgradeRequest struct {
		UserID               *uint  `json:"user_id"`
		StripeSubscriptionID string `json:"stripe_subscription_id" binding:"required"`
		ProductID            uint   `json:"product_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&downgradeRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retrieve the user from the database
	var user models.User
	if err := config.DB.Where("id = ?", downgradeRequest.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Cancel the Stripe subscription
	_, err := sub.Cancel(downgradeRequest.StripeSubscriptionID, &stripe.SubscriptionCancelParams{
		InvoiceNow: stripe.Bool(true),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel subscription"})
		return
	}

	// Create a free subscription in the database
	freeSubscription := models.Subscription{
		UserID:               downgradeRequest.UserID,
		StripeSubscriptionID: "free",
		SubscriptionStatus:   "free",
		Quantity:             1,
		Active:               stripe.Bool(true),
		PriceId:              "free",
		ProductID:            downgradeRequest.ProductID,
	}
	if err := config.DB.Create(&freeSubscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create free subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription downgraded to free"})
}

// downgrade organization subscription
func DowngradeOrganizationSubscription(c *gin.Context) {
	var downgradeRequest struct {
		OrganizationID       uint   `json:"organization_id" binding:"required"`
		StripeSubscriptionID string `json:"stripe_subscription_id" binding:"required"`
		ProductID            uint   `json:"product_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&downgradeRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retrieve the organization from the database
	var organization models.Organization
	if err := config.DB.Where("id = ?", downgradeRequest.OrganizationID).First(&organization).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	// Cancel the Stripe subscription
	_, err := sub.Cancel(downgradeRequest.StripeSubscriptionID, &stripe.SubscriptionCancelParams{
		InvoiceNow: stripe.Bool(true),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel subscription"})
		return
	}

	// Create a free subscription in the database
	freeSubscription := models.Subscription{
		OrganizationID:       &downgradeRequest.OrganizationID,
		StripeSubscriptionID: "free",
		SubscriptionStatus:   "free",
		Quantity:             1,
		Active:               stripe.Bool(false),
		PriceId:              "free",
		ProductID:            downgradeRequest.ProductID,
	}
	if err := config.DB.Create(&freeSubscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create free subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription downgraded to free"})
}
