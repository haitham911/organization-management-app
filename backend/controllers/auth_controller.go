package controllers

import (
	"net/http"
	"organization-management-app/config"
	"organization-management-app/form"
	"organization-management-app/models"
	"organization-management-app/services"
	"organization-management-app/utils"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InviteUser godoc
// @Summary Invite a user to an organization with role Admin or Member
// @Description Invite a user to an organization by sending a magic link to their email
// @Tags users
// @Accept json
// @Produce json
// @Param invite body form.InviteUserRequest true "Invite User"
// @Success 200 {object} form.MessageResponse
// @Failure 400 {object} form.ErrorResponse
// @Failure 404 {object} form.ErrorResponse
// @Failure 500 {object} form.ErrorResponse
// @Router /users/invite [post]
func InviteUser(c *gin.Context) {
	var request form.InviteUserRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, form.ErrorResponse{Error: err.Error()})
		return
	}
	if request.Role != "Admin" && request.Role != "Member" {
		c.JSON(http.StatusBadRequest, form.ErrorResponse{Error: "invalid role"})
		return
	}
	token, err := utils.GenerateMagicLinkToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, form.ErrorResponse{Error: "Failed to generate token"})
		return
	}

	magicLinkExpiry := time.Now().Add(24 * time.Hour)

	user := models.User{Email: request.Email, MagicLinkToken: token, MagicLinkExpiry: magicLinkExpiry}
	config.DB.FirstOrCreate(&user, "email = ?", user.Email)

	var organization models.Organization
	if err := config.DB.First(&organization, request.OrganizationID).Error; err != nil {
		c.JSON(http.StatusNotFound, form.ErrorResponse{Error: "Organization not found"})
		return
	}

	organizationUser := models.OrganizationUser{
		UserID:         user.ID,
		OrganizationID: organization.ID,
		Role:           request.Role, // Default role, you can change it as needed
	}

	if err := config.DB.Create(&organizationUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, form.ErrorResponse{Error: "Failed to create organization user"})
		return
	}
	link := os.Getenv("FRONT_INVITE_URL") + token

	if err := services.SendEmail(user.Email, "Organization Invitation", "You have been invited to join an organization. Click the link to accept: "+link, link); err != nil {
		c.JSON(http.StatusInternalServerError, form.ErrorResponse{Error: "Failed to send invite email. email Services error" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, form.MessageResponse{Message: "Invitation sent"})
}

func VerifyMagicLink(c *gin.Context) {
	token := c.Query("token")
	var user models.User
	if err := config.DB.Where("magic_link_token = ? AND magic_link_expiry > ?", token, time.Now()).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	jwtToken, err := utils.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	config.DB.Model(&user).Updates(models.User{MagicLinkToken: "", MagicLinkExpiry: time.Time{}})

	c.JSON(http.StatusOK, gin.H{"token": jwtToken})
}
func Login(c *gin.Context) {
	var loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Preload("Organizations.Subscriptions").Where("email = ?", loginRequest.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if !utils.CheckPasswordHash(loginRequest.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Check if the user belongs to any organization with an active subscription
	hasActiveSubscription := false
	for _, org := range user.Organizations {
		for _, sub := range org.Subscriptions {
			if *sub.Active {
				hasActiveSubscription = true
				break
			}
		}
		if hasActiveSubscription {
			break
		}
	}

	if !hasActiveSubscription {
		c.JSON(http.StatusForbidden, gin.H{"error": "User does not belong to an organization with an active subscription"})
		return
	}

	// Generate token
	token, err := utils.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// SignUpWithMagicLink godoc
// @Summary Sign up with a magic link
// @Description Sign up with a magic link sent to the user's email
// @Tags users
// @Accept json
// @Produce json
// @Param email body form.EmailRequest true "Email"
// @Success 200 {object} form.MessageResponse
// @Failure 400 {object} form.ErrorResponse
// @Failure 500 {object} form.ErrorResponse
// @Router /users/signup-magic-link [post]
func SignUpWithMagicLink(c *gin.Context) {
	var request form.EmailRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, form.ErrorResponse{Error: err.Error()})
		return
	}

	magicLinkToken := uuid.New().String()
	magicLinkExpiry := time.Now().Add(15 * time.Minute)
	active := false
	user := models.User{
		Email:           request.Email,
		MagicLinkToken:  magicLinkToken,
		MagicLinkExpiry: magicLinkExpiry,
		Active:          &active,
	}

	if err := config.DB.Debug().FirstOrCreate(&user, "email=?", request.Email).Error; err != nil {
		if strings.Contains(err.Error(), "email") {
			c.JSON(http.StatusBadRequest, form.ErrorResponse{Error: "user email already exist"})
			return
		}

		c.JSON(http.StatusInternalServerError, form.ErrorResponse{Error: "Failed to create user"})
		return
	}
	if *user.Active {
		c.JSON(http.StatusBadRequest, form.ErrorResponse{Error: "user email already exist and email verified"})
		return
	}
	baseUrl := utils.RemoveLastSlash(os.Getenv("SERVER_URL"))
	subject := "Complete your sign up"
	plainTextContent := "Click the link to verify your email: " + baseUrl + "/complete-signup?token=" + magicLinkToken
	htmlContent := "<strong>Click the link to verify your email: <a href=\"" + baseUrl + "/complete-signup?token=" + magicLinkToken + "\">Complete Signup</a></strong>"

	if err := services.SendEmail(user.Email, subject, plainTextContent, htmlContent); err != nil {
		c.JSON(http.StatusInternalServerError, form.ErrorResponse{Error: "Failed to send sing up email. email Services error" + err.Error()})

		return
	}

	c.JSON(http.StatusOK, form.MessageResponse{Message: "Signup email sent"})
}

// CompleteSignup godoc
// @Summary Complete the signup process
// @Description Complete the signup process using the magic link
// @Tags users
// @Accept json
// @Produce json
// @Param token query string true "Magic Link Token"
// @Param password body form.PasswordRequest true "Password"
// @Success 200 {object} form.MessageResponse
// @Failure 400 {object} form.ErrorResponse
// @Failure 500 {object} form.ErrorResponse
// @Router /users/complete-signup [post]
func CompleteSignup(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, form.ErrorResponse{Error: "Token is required"})
		return
	}

	var user models.User
	if err := config.DB.Where("magic_link_token = ? AND magic_link_expiry > ?", token, time.Now()).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, form.ErrorResponse{Error: "Invalid or expired token"})
		return
	}

	user.MagicLinkToken = ""
	user.MagicLinkExpiry = time.Time{}
	active := true
	user.Active = &active

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, form.ErrorResponse{Error: "Failed to complete signup"})
		return
	}

	c.JSON(http.StatusOK, form.MessageResponse{Message: "Signup complete"})
}

// LoginWithMagicLink godoc
// @Summary Login with a magic link
// @Description Login with a magic link sent to the user's email
// @Tags users
// @Accept json
// @Produce json
// @Param email body form.EmailRequest true "Email"
// @Success 200 {object} form.MessageResponse
// @Failure 400 {object} form.ErrorResponse
// @Failure 500 {object} form.ErrorResponse
// @Router /users/login-magic-link [post]
func LoginWithMagicLink(c *gin.Context) {
	var request form.EmailRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, form.ErrorResponse{Error: err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", request.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, form.ErrorResponse{Error: "User not found"})
		return
	}

	magicLinkToken := uuid.New().String()
	magicLinkExpiry := time.Now().Add(15 * time.Minute)

	user.MagicLinkToken = magicLinkToken
	user.MagicLinkExpiry = magicLinkExpiry

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, form.ErrorResponse{Error: "Failed to update user"})
		return
	}

	subject := "Login to your account"
	baseUrl := utils.RemoveLastSlash(os.Getenv("SERVER_URL"))

	plainTextContent := "Click the link to log in: " + baseUrl + "/login?token=" + magicLinkToken
	htmlContent := "<strong>Click the link to log in: <a href=\"" + baseUrl + "/login?token=" + magicLinkToken + "\">Login</a></strong>"

	if err := services.SendEmail(user.Email, subject, plainTextContent, htmlContent); err != nil {
		c.JSON(http.StatusInternalServerError, form.ErrorResponse{Error: "Failed to send login email. email Services error" + err.Error()})

		return
	}

	c.JSON(http.StatusOK, form.MessageResponse{Message: "Login email sent"})
}

// MagicLinkLogin godoc
// @Summary Complete the login process using the magic link
// @Description Complete the login process using the magic link
// @Tags users
// @Accept json
// @Produce json
// @Param token query string true "Magic Link Token"
// @Success 200 {object} form.MessageResponse
// @Failure 400 {object} form.ErrorResponse
// @Failure 500 {object} form.ErrorResponse
// @Router /users/login [post]
func MagicLinkLogin(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, form.ErrorResponse{Error: "Token is required"})
		return
	}

	var user models.User
	if err := config.DB.Where("magic_link_token = ? AND magic_link_expiry > ?", token, time.Now()).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, form.ErrorResponse{Error: "Invalid or expired token"})
		return
	}

	// Clear the magic link token and expiry
	user.MagicLinkToken = ""
	user.MagicLinkExpiry = time.Time{}

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, form.ErrorResponse{Error: "Failed to complete login"})
		return
	}

	//Generate JWT token or session here and return to the user
	token, err := utils.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, form.ErrorResponse{Error: "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, form.TokenResponse{Token: token})
}

// GetUserWithRoles godoc
// @Summary Get user information with roles in organizations
// @Description Get user information with roles in organizations
// @Tags users
// @Accept json
// @Produce json
// @Param user_id path uint true "User ID"
// @Failure 400 {object} form.ErrorResponse
// @Failure 404 {object} form.ErrorResponse
// @Router /users/{user_id}/roles [get]
func GetUserWithRoles(c *gin.Context) {
	userID := c.Param("user_id")

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, form.ErrorResponse{Error: "User not found"})
		return
	}

	var organizationUsers []models.OrganizationUser
	if err := config.DB.Preload("Organization").Preload("User").Where("user_id = ?", userID).Find(&organizationUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, form.ErrorResponse{Error: "Failed to retrieve organization users"})
		return
	}

	var userWithRoles []models.UserWithRoles
	for _, orgUser := range organizationUsers {
		userWithRoles = append(userWithRoles, models.UserWithRoles{
			Role:         orgUser.Role,
			Organization: orgUser.Organization,
		})
	}

	c.JSON(http.StatusOK, userWithRoles)
}
