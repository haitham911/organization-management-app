package controllers

import (
	"fmt"
	"log"
	"net/http"
	"organization-management-app/config"
	"organization-management-app/form"
	"organization-management-app/models"
	"organization-management-app/services"
	"organization-management-app/utils"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v72/customer"
	"gorm.io/gorm"

	stripe "github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/invoice"
	"github.com/stripe/stripe-go/v72/sub"
)

type CreateOrganizationReq struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
}

// CreateOrganization godoc
// @Summary create organization
// @Description create organization
// @Tags organizations
// @Accept json
// @Produce json
// @Security Bearer
// @Param data body CreateOrganizationReq true "body"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /organization [post]
func CreateOrganization(c *gin.Context) {
	profile, err := utils.GetProfileFromGinCtx(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var req CreateOrganizationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Create Stripe Customer
	params := &stripe.CustomerParams{
		Name:  stripe.String(req.Name),
		Email: stripe.String(req.Email), // Assuming the organization has an email field
	}
	stripeCustomer, err := customer.New(params)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Stripe customer"})
		return
	}

	org := models.Organization{
		Name:             req.Name,
		Email:            req.Email,
		StripeCustomerID: stripeCustomer.ID,
	}

	result := config.DB.Create(&org)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, result.Error.Error())
		return
	}
	userOrg := models.OrganizationUser{
		UserID:         profile.UserID,
		OrganizationID: org.ID,
		Role:           "Admin",
	}

	config.DB.Create(&userOrg)
	newJwt, err := utils.GenerateToken(profile.UserID)
	if err != nil {
		log.Println(err)
	}
	type JsonResponse struct {
		Data  interface{} `json:"data"`
		Token string      `json:"token"`
	}
	c.JSON(http.StatusOK, JsonResponse{Data: org, Token: newJwt})
}

// GetOrganizationsUsers users godoc
// @Summary Get the Organizations users
// @Description  Get the Organizations users
// @Tags organizations
// @Accept json
// @Produce json
// @Security Bearer
// @Param orgId query int  true "Organization Id"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /organizations/users [get]
func GetOrganizationsUsers(c *gin.Context) {
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

	var organization models.Organization
	if err := config.DB.Preload("Users", func(db *gorm.DB) *gorm.DB {
		return db.Select("users.id, users.name, users.email")
	}).First(&organization, orgId).Error; err != nil {
		c.JSON(http.StatusNotFound, form.ErrorResponse{Error: "Organization not found"})
		return
	}
	c.JSON(http.StatusOK, organization)
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
	if err := config.DB.Model(&models.OrganizationUser{}).Where("organization_id = ? AND stripe_subscription_id = ?", request.OrganizationID, request.StripeSubscriptionID).Count(&totalUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count users"})
		return
	}

	canAddMoreSubscriptions := totalUsers < int64(subscription.Quantity)

	c.JSON(http.StatusOK, gin.H{"can_add_more_subscriptions": canAddMoreSubscriptions})
}

// GetOrganizationSubscriptionInfo godoc
// @Summary Get the number of members and remaining subscriptions for an organization
// @Description Retrieve the number of members in an organization and how many subscriptions are left
// @Tags organizations
// @Accept json
// @Produce json
// @Security Bearer
// @Param orgId query int  true "Organization Id"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /organizations/subscription-info [get]
func GetOrganizationSubscriptionInfo(c *gin.Context) {

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
	// Find the organization's subscription
	var subscription models.Subscription
	if err := config.DB.Where("organization_id = ?", orgId).First(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subscription"})
		return
	}

	// Count the number of members in the organization
	var totalUsers int64
	if err := config.DB.Model(&models.OrganizationUser{}).Where("organization_id = ?", orgId).Count(&totalUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count members"})
		return
	}
	// Retrieve the subscription from Stripe
	stripe.Key = config.GetStripeSecretKey()
	subscriptionInfo, err := sub.Get(subscription.StripeSubscriptionID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Stripe subscription"})
		return
	}

	// Calculate remaining subscriptions
	remainingSubscriptions := subscription.Quantity - int(totalUsers)

	c.JSON(http.StatusOK, gin.H{
		"total_members":           totalUsers,
		"remaining_subscriptions": remainingSubscriptions,
		"subscription_info":       subscriptionInfo,
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
// @Security Bearer
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
	if err := config.DB.Where("user_id = ? AND organization_id = ?", request.UserID, request.OrganizationID).Delete(&models.OrganizationUser{}).Error; err != nil {
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
	Email                string `json:"email" binding:"required"`
	StripeSubscriptionID string `json:"stripe_subscription_id" binding:"required"`
	Role                 string `json:"role"`
}

// SendInvite godoc
// @Summary Send an invite to a user
// @Description Send an invite to a user to join the organization
// @Tags subscriptions
// @Accept json
// @Produce json
// @Security Bearer
// @Param orgId query int  true "Organization Id"
// @Param data body InviteReq true "body"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /subscriptions/send-invite [post]
func SendInvite(c *gin.Context) {
	var inviteRequest InviteReq
	if err := c.ShouldBindJSON(&inviteRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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
	var org models.Organization
	result := config.DB.Where("id=?", orgId).First(&org)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization not found"})
		c.Abort()
		return
	}
	var sub models.Subscription
	result = config.DB.Where("organization_id=? and stripe_subscription_id=? ", orgId, inviteRequest.StripeSubscriptionID).First(&sub)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization subscription not found"})
		c.Abort()
		return
	}
	invite := models.UserInvite{
		Email:                inviteRequest.Email,
		OrganizationID:       uint(orgId),
		InviteToken:          uuid.New().String(),
		IsAccepted:           false,
		StripeSubscriptionID: inviteRequest.StripeSubscriptionID,
		Role:                 inviteRequest.Role,
	}

	if err := config.DB.Create(&invite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invite"})
		return
	}
	subject := fmt.Sprintf("Invitation from %s", org.Name)
	baseUrl := utils.RemoveLastSlash(os.Getenv("FRONT_URL"))

	plainTextContent := "Click the link to to accept Invitation: " + baseUrl + "/accept-invite?token=" + invite.InviteToken
	htmlContent := "<strong>Click the link to to accept Invitation: <a href=\"" + baseUrl + "/accept-invite?token=" + invite.InviteToken + "\">Login</a></strong>"

	if err := services.SendEmail(invite.Email, subject, plainTextContent, htmlContent); err != nil {
		c.JSON(http.StatusInternalServerError, form.ErrorResponse{Error: "Failed to send login email. email Services error" + err.Error()})

		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Invite sent successfully"})
}

type AcceptInviteReq struct {
	InviteToken string `json:"invite_token" binding:"required"`
	Name        string `json:"name" binding:"required"`
}

// AcceptInvite godoc
// @Summary Accept an invite to join the organization
// @Description Accept an invite and create a user in the organization
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param data body AcceptInviteReq true "body"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /subscription/accept-invite [post]
func AcceptInvite(c *gin.Context) {
	var acceptRequest struct {
		InviteToken string `json:"invite_token" binding:"required"`
		Name        string `json:"name" binding:"required"`
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

	user := models.User{
		Email:            invite.Email,
		Name:             acceptRequest.Name,
		StripeCustomerID: "",
	}
	// Find the organization's subscription
	var subscription models.Subscription
	if err := config.DB.Where("organization_id = ?", invite.OrganizationID).First(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subscription"})
		return
	}
	tx := config.DB.Begin()

	if err := tx.FirstOrCreate(&user).Where("email=?", invite.Email).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Create OrganizationUser association
	userOrg := models.OrganizationUser{
		UserID:               user.ID,
		OrganizationID:       invite.OrganizationID,
		StripeSubscriptionID: subscription.StripeSubscriptionID,
		Role:                 invite.Role,
	}

	if err := tx.Create(&userOrg).Error; err != nil {
		tx.Rollback()

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user-organization association"})
		return
	}
	// Mark the invite as accepted
	invite.IsAccepted = true
	if err := tx.Save(&invite).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update invite"})
		return
	}
	// Retrieve the subscription from Stripe
	stripe.Key = config.GetStripeSecretKey()
	// Retrieve the subscription from Stripe to get the correct subscription item ID
	stripeSubscription, err := sub.Get(subscription.StripeSubscriptionID, nil)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve Stripe subscription"})
		return
	}

	if len(stripeSubscription.Items.Data) == 0 {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No subscription items found in Stripe subscription"})
		return
	}

	// Update the subscription to add a seat (increment quantity)
	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:       stripe.String(stripeSubscription.Items.Data[0].ID),
				Quantity: stripe.Int64(int64(stripeSubscription.Quantity + 1)),
			},
		},
		ProrationBehavior: stripe.String("create_prorations"),
	}

	updatedSubscription, err := sub.Update(subscription.StripeSubscriptionID, params)
	if err != nil {
		tx.Rollback()
		msg := fmt.Sprintf("updatedSubscription error : %v", err)
		println(msg)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Stripe subscription"})
		return
	}
	tx.Commit()

	subscription.Quantity = int(updatedSubscription.Items.Data[0].Quantity)
	if err := config.DB.Save(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription in database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invite accepted successfully", "user": user})
}

type GetProratedCostReq struct {
	OrganizationID       uint   `json:"organization_id" binding:"required"`
	StripeSubscriptionID string `json:"stripe_subscription_id" binding:"required"`
	SeatCount            int    `json:"seat_count" binding:"required"`
}

func GetProratedCost(c *gin.Context) {
	var request GetProratedCostReq

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

	if len(subscription.Items.Data) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No subscription items found in Stripe subscription"})
		return
	}

	subscriptionItemID := subscription.Items.Data[0].ID

	// Form the parameters for the upcoming invoice
	invoiceParams := &stripe.InvoiceParams{
		Customer:     stripe.String(subscription.Customer.ID),
		Subscription: stripe.String(subscription.ID),
		SubscriptionItems: []*stripe.SubscriptionItemsParams{
			{
				ID:       stripe.String(subscriptionItemID),
				Quantity: stripe.Int64(int64(request.SeatCount)),
			},
		},
		SubscriptionProrationBehavior: stripe.String("create_prorations"),
	}

	prorationInvoice, err := invoice.GetNext(invoiceParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate prorated cost"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"prorated_cost": prorationInvoice.Total / 100, "invoice": prorationInvoice})
}

// DisableUser godoc
// @Summary Disable a user and remove their seat from the organization
// @Description Disable a user and remove their seat from the organization without deleting the user
// @Tags subscriptions
// @Accept json
// @Produce json
// @Security Bearer
// @Param orgId query int  true "Organization Id"
// @Param request body DisableUserRequest true "Disable User"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /subscriptions/disable-user [post]
func DisableUser(c *gin.Context) {
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
	if err := config.DB.Where("id = ?", request.UserID).First(&user).Error; err != nil {
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

	// Disable the user
	*user.Active = false
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable user"})
		return
	}

	// Update the subscription details in the database
	subscription.Quantity = int(updatedSubscription.Items.Data[0].Quantity)
	if err := config.DB.Save(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription in database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User disabled and seat removed successfully"})
}

// GetOrganizationsUsers Pending  users godoc
// @Summary Get the Organizations users
// @Description  Get the Organizations users
// @Tags organizations
// @Accept json
// @Produce json
// @Security Bearer
// @Param orgId query int  true "Organization Id"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /organizations/pending [get]
func GetOrganizationsUsersPending(c *gin.Context) {
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

	var userInvite models.UserInvite
	config.DB.Where("organization_id = ? and is_accepted = false", orgId).First(&userInvite)

	c.JSON(http.StatusOK, userInvite)
}

// GetOrganizations users godoc
// @Summary Get the Organizations 
// @Description  Get the Organizations 
// @Tags organizations
// @Accept json
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /all/organization [get]
func GetOrganizations(c *gin.Context) {
	
	var organization []models.Organization
	if err := config.DB.Find(&organization).Error; err != nil {
		c.JSON(http.StatusNotFound, form.ErrorResponse{Error: "Organization not found"})
		return
	}
	c.JSON(http.StatusOK, organization)
}