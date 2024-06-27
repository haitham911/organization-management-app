package models

import (
	"time"

	"gorm.io/gorm"
)

type Organization struct {
	gorm.Model
	Name             string         `json:"name"`
	Email            string         `json:"email" gorm:"unique"`
	StripeCustomerID string         `json:"stripe_customer_id"` // Stripe Customer ID for billing
	Users            []User         `gorm:"many2many:organization_users;"`
	Subscriptions    []Subscription `json:"subscriptions"`
}
type User struct {
	gorm.Model
	Name            string         `json:"name"`
	Email           string         `json:"email" gorm:"unique"`
	Password        string         `json:"password"`
	Role            string         `json:"role"` // Admin or User
	MagicLinkToken  string         `json:"magic_link_token"`
	MagicLinkExpiry time.Time      `json:"magic_link_expiry"`
	Organizations   []Organization `gorm:"many2many:organization_users;"`
}

type Product struct {
	gorm.Model
	Name        string `json:"name"`
	PriceID     string `json:"price_id"`     // Stripe Price ID
	PriceAmount int64  `json:"price_amount"` // Price per user in cents
}
type Subscription struct {
	gorm.Model
	OrganizationID       uint
	PriceId              string
	StripeSubscriptionID string `json:"stripe_subscription_id"`
	Quantity             int    `json:"quantity"` // Number of users/seats
	Active               bool   `json:"active"`   // Subscription active status
	SubscriptionStatus   string
	ProductID            uint `json:"product_id" binding:"required"`
}
type UserOrganization struct {
	gorm.Model
	UserID               uint   `json:"user_id"`
	OrganizationID       uint   `json:"organization_id"`
	StripeSubscriptionID string `json:"stripe_subscription_id"`
}
