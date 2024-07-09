package controllers

import (
	"log"
	"net/http"
	"organization-management-app/config"
	"organization-management-app/models"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/invoice"
	"github.com/stripe/stripe-go/sub"
	"golang.org/x/crypto/bcrypt"
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

// Check if an organization can add more subscriptions
func CanAddMoreSubscriptions(c *gin.Context) {
	var request struct {
		OrganizationID       uint   `json:"organization_id" binding:"required"`
		StripeSubscriptionID string `json:"stripe_subscription_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var subscription models.Subscription
	if err := config.DB.Where("organization_id = ? AND stripe_subscription_id = ?", request.OrganizationID, request.StripeSubscriptionID).First(&subscription).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	var totalUsers int64
	if err := config.DB.Model(&models.UserOrganization{}).Where("organization_id = ? AND stripe_subscription_id = ?", request.OrganizationID, request.StripeSubscriptionID).Count(&totalUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count users"})
		return
	}

	canAddMoreSubscriptions := totalUsers < int64(subscription.Quantity)

	c.JSON(http.StatusOK, gin.H{"can_add_more_subscriptions": canAddMoreSubscriptions})
}

// GetOrganizationSubscriptionInfo godoc
// @Summary Get the number of members and remaining subscriptions for an organization
// @Description Retrieve the number of members in an organization and how many subscriptions are left
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param organization_id query uint true "Organization ID"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /subscriptions/organization-info [get]
func GetOrganizationSubscriptionInfo(c *gin.Context) {
	var query struct {
		OrganizationID uint `form:"organization_id" binding:"required"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the organization's subscription
	var subscription models.Subscription
	if err := config.DB.Where("organization_id = ?", query.OrganizationID).First(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subscription"})
		return
	}

	// Count the number of members in the organization
	var totalUsers int64
	if err := config.DB.Model(&models.UserOrganization{}).Where("organization_id = ?", query.OrganizationID).Count(&totalUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count members"})
		return
	}

	// Calculate remaining subscriptions
	remainingSubscriptions := subscription.Quantity - int(totalUsers)

	c.JSON(http.StatusOK, gin.H{
		"total_members":           totalUsers,
		"remaining_subscriptions": remainingSubscriptions,
	})
}

type RemoveUserReq struct {
	UserID         uint `json:"user_id" binding:"required"`
	OrganizationID uint `json:"organization_id" binding:"required"`
}

// RemoveUser godoc
// @Summary Remove a user from the organization
// @Description Remove a user from an organization and update the subscription
// @Tags subscriptions
// @Accept json
// @Produce json
//
// @Param request body RemoveUserReq true "Remove User"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /subscriptions/remove-user [post]
func RemoveUser(c *gin.Context) {
	var request struct {
		UserID         uint `json:"user_id" binding:"required"`
		OrganizationID uint `json:"organization_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retrieve the user's details
	var user models.User
	if err := config.DB.Where("id = ? AND organizations.id = ?", request.UserID, request.OrganizationID).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	// Retrieve the subscription from Stripe
	stripe.Key = config.GetStripeSecretKey()
	var subscription models.Subscription
	if err := config.DB.Where("organization_id = ?", request.OrganizationID).First(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subscription"})
		return
	}

	// Update the subscription to remove a seat (decrement quantity)
	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:       stripe.String(subscription.StripeSubscriptionID),
				Quantity: stripe.Int64(int64(subscription.Quantity - 1)),
			},
		},
		ProrationBehavior: stripe.String("create_prorations"),
	}

	updatedSubscription, err := sub.Update(subscription.StripeSubscriptionID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Stripe subscription"})
		return
	}

	// Remove the user's association with the organization
	if err := config.DB.Where("user_id = ? AND organization_id = ?", request.UserID, request.OrganizationID).Delete(&models.UserOrganization{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user from organization"})
		return
	}

	// Delete the user
	if err := config.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	// Update the subscription details in the database
	subscription.Quantity = int(updatedSubscription.Items.Data[0].Quantity)
	if err := config.DB.Save(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription in database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User removed successfully"})
}

type AddSeatReq struct {
	UserID               uint   `json:"user_id"`
	OrganizationID       uint   `json:"organization_id"`
	StripeSubscriptionID string `json:"stripe_subscription_id" binding:"required"`
}

// AddSeat godoc
// @Summary Add a seat to the subscription
// @Description Add a user seat to an existing subscription with prorated billing
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body AddSeatReq true "Add Seat"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /subscriptions/add-seat [post]
func AddSeat(c *gin.Context) {
	var seatRequest struct {
		UserID               uint   `json:"user_id"`
		OrganizationID       uint   `json:"organization_id"`
		StripeSubscriptionID string `json:"stripe_subscription_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&seatRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retrieve the subscription from Stripe
	stripe.Key = config.GetStripeSecretKey()
	subscription, err := sub.Get(seatRequest.StripeSubscriptionID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Stripe subscription"})
		return
	}

	// Update the subscription to add a seat (increment quantity)
	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:       stripe.String(subscription.Items.Data[0].ID),
				Quantity: stripe.Int64(subscription.Items.Data[0].Quantity + 1),
			},
		},
		ProrationBehavior: stripe.String("create_prorations"),
	}

	updatedSubscription, err := sub.Update(subscription.ID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Stripe subscription"})
		return
	}

	// Update the subscription details in the database
	var dbSubscription models.Subscription
	if err := config.DB.Where("stripe_subscription_id = ?", seatRequest.StripeSubscriptionID).First(&dbSubscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find subscription in database"})
		return
	}

	dbSubscription.Quantity = int(updatedSubscription.Items.Data[0].Quantity)
	if err := config.DB.Save(&dbSubscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription in database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Seat added successfully", "subscription": updatedSubscription})
}

type InviteReq struct {
	Email          string `json:"email" binding:"required"`
	OrganizationID uint   `json:"organization_id" binding:"required"`
}

// SendInvite godoc
// @Summary Send an invite to a user
// @Description Send an invite to a user to join the organization
// @Tags users
// @Accept json
// @Produce json
//
// @Param data body InviteReq true "body"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /users/send-invite [post]
func SendInvite(c *gin.Context) {
	var inviteRequest struct {
		Email          string `json:"email" binding:"required"`
		OrganizationID uint   `json:"organization_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&inviteRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invite := models.UserInvite{
		Email:          inviteRequest.Email,
		OrganizationID: inviteRequest.OrganizationID,
		InviteToken:    uuid.New().String(),
		IsAccepted:     false,
	}

	if err := config.DB.Create(&invite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invite"})
		return
	}

	// TODO: Send invite email to user with invite.InviteToken

	c.JSON(http.StatusOK, gin.H{"message": "Invite sent successfully"})
}

type AcceptInviteReq struct {
	InviteToken string `json:"invite_token" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

// AcceptInvite godoc
// @Summary Accept an invite to join the organization
// @Description Accept an invite and create a user in the organization
// @Tags users
// @Accept json
// @Produce json
// @Param data body AcceptInviteReq true "body"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /users/accept-invite [post]
func AcceptInvite(c *gin.Context) {
	var acceptRequest struct {
		InviteToken string `json:"invite_token" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Password    string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&acceptRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var invite models.UserInvite
	if err := config.DB.Where("invite_token = ? AND is_accepted = ?", acceptRequest.InviteToken, false).First(&invite).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or already accepted invite token"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(acceptRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Email:            invite.Email,
		Name:             acceptRequest.Name,
		Password:         string(hashedPassword),
		StripeCustomerID: "",
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Create UserOrganization association
	userOrg := models.UserOrganization{
		UserID:         user.ID,
		OrganizationID: invite.OrganizationID,
	}

	if err := config.DB.Create(&userOrg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user-organization association"})
		return
	}

	// Update the subscription to add a seat (increment quantity)
	var subscription models.Subscription
	if err := config.DB.Where("organization_id = ?", invite.OrganizationID).First(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subscription"})
		return
	}

	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:       stripe.String(subscription.StripeSubscriptionID),
				Quantity: stripe.Int64(int64(subscription.Quantity + 1)),
			},
		},
		ProrationBehavior: stripe.String("create_prorations"),
	}

	updatedSubscription, err := sub.Update(subscription.StripeSubscriptionID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Stripe subscription"})
		return
	}

	subscription.Quantity = int(updatedSubscription.Items.Data[0].Quantity)
	if err := config.DB.Save(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription in database"})
		return
	}

	// Mark the invite as accepted
	invite.IsAccepted = true
	if err := config.DB.Save(&invite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update invite"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invite accepted successfully", "user": user})
}

type GetProratedCostReq struct {
	OrganizationID       uint   `json:"organization_id" binding:"required"`
	StripeSubscriptionID string `json:"stripe_subscription_id" binding:"required"`
	SeatCount            int    `json:"seat_count" binding:"required"`
}

// GetProratedCost godoc
// @Summary Get the prorated cost for adding a seat to a subscription
// @Description Get the prorated cost for adding a seat to a subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param data body GetProratedCostReq true "body"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /subscriptions/prorated-cost [post]
func GetProratedCost(c *gin.Context) {
	var request struct {
		OrganizationID       uint   `json:"organization_id" binding:"required"`
		StripeSubscriptionID string `json:"stripe_subscription_id" binding:"required"`
		SeatCount            int    `json:"seat_count" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retrieve the subscription from Stripe
	stripe.Key = config.GetStripeSecretKey()
	subscription, err := sub.Get(request.StripeSubscriptionID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Stripe subscription"})
		return
	}

	// Calculate the prorated cost
	items := []*stripe.SubscriptionItemsParams{
		{
			ID:       stripe.String(subscription.Items.Data[0].ID),
			Quantity: stripe.Int64(subscription.Items.Data[0].Quantity + int64(request.SeatCount)),
		},
	}

	invoiceParams := &stripe.InvoiceParams{
		Customer:                      stripe.String(subscription.Customer.ID),
		SubscriptionItems:             items,
		SubscriptionProrationBehavior: stripe.String("create_prorations"),
	}

	prorationInvoice, err := invoice.GetNext(invoiceParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate prorated cost"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"prorated_cost": prorationInvoice.Total, "invoice": prorationInvoice})
}
