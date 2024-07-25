package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"organization-management-app/config"
	"organization-management-app/models"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func getJWTSecret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

type Claims struct {
	UserID             uint   `json:"user_id"`
	Email              string `json:"email"`
	UsersSubscriptions []models.Subscription
	Organizations      []models.UserWithRoles `json:"organizations"`

	jwt.StandardClaims
}

// `json:"user_subscriptions"`
func GenerateMagicLinkToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GenerateToken(userId uint) (string, error) {
	var userToken models.User
	err := config.DB.Where("id = ?", userId).First(&userToken).Preload("Subscriptions").Error
	if err != nil {
		return "", err
	}
	var organizationUsers []models.OrganizationUser
	err = config.DB.Preload("Organization").Where("user_id = ?", userToken.ID).Find(&organizationUsers).Error
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
		UserID:             userToken.ID,
		Email:              userToken.Email,
		UsersSubscriptions: userToken.Subscriptions,
		Organizations:      userWithRoles,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 336).Unix(),
		},
	})
	return token.SignedString(getJWTSecret())
}

func ParseToken(tokenString string, claim *Claims) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, claim, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return getJWTSecret(), nil
	})
}
func GetProfileFromGinCtx(c *gin.Context) (*Claims, error) {
	profile, ok := c.Get("user")
	if !ok {
		return nil, errors.New("invalid user")
	}
	me, err := GetProfile(profile)
	if err != nil {
		return me, err
	}
	return me, nil
}
func GetProfile(profile interface{}) (*Claims, error) {
	jsonStr, err := json.Marshal(profile)
	if err != nil {
		return nil, err
	}
	var session Claims
	if err = json.Unmarshal(jsonStr, &session); err != nil {
		return nil, err
	}
	return &session, nil
}
