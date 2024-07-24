package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"organization-management-app/config"
	"organization-management-app/models"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gorm.io/gorm"
)

var jwtSecret = []byte("your_secret_key")

type Claims struct {
	UserID             uint   `json:"user_id"`
	Email              string `json:"email"`
	UsersSubscriptions []models.Subscription
	Organizations      []models.UserWithRoles `json:"organizations"`

	jwt.StandardClaims
}

func GenerateMagicLinkToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GenerateToken(user models.User) (string, error) {
	var userToken models.User
	err := config.DB.Where("email = ?", user.Email).First(&userToken).Preload("Subscriptions").Error
	if err != nil {
		return "", err
	}
	var organizationUsers []models.OrganizationUser
	err = config.DB.Preload("Organization").Where("user_id = ?", user.ID).Find(&organizationUsers).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	var userWithRoles []models.UserWithRoles
	for _, orgUser := range organizationUsers {
		userWithRoles = append(userWithRoles, models.UserWithRoles{
			Role:                    orgUser.Role,
			OrgStripeSubscriptionID: orgUser.StripeSubscriptionID,
			Organization:            orgUser.Organization,
		})
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:             user.ID,
		Email:              user.Email,
		UsersSubscriptions: user.Subscriptions,
		Organizations:      userWithRoles,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	})
	return token.SignedString(jwtSecret)
}

func ParseToken(tokenString string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
}
