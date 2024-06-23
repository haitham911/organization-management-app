package controllers

import (
	"net/http"
	"organization-management-app/config"
	"organization-management-app/models"
	"organization-management-app/services"
	"organization-management-app/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func InviteUser(c *gin.Context) {
	var request struct {
		Email          string `json:"email"`
		OrganizationID uint   `json:"organization_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := utils.GenerateMagicLinkToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	magicLinkExpiry := time.Now().Add(24 * time.Hour)

	user := models.User{Email: request.Email, MagicLinkToken: token, MagicLinkExpiry: magicLinkExpiry}
	config.DB.FirstOrCreate(&user, "email = ?", user.Email)
	var organization models.Organization
	if err := config.DB.First(&organization, request.OrganizationID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}
	config.DB.Model(&user).Association("Organizations").Append(&organization)

	link := "http://yourfrontend.com/magic-link?token=" + token

	if err := services.SendEmail(request.Email, link); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invitation sent"})
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
