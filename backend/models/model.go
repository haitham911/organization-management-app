package models

import (
	"database/sql"
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
	ID               uint `gorm:"primarykey"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        sql.NullTime   `gorm:"index"`
	Usage            int            `json:"usage" gorm:"default:0"`
	StripeCustomerID string         `json:"stripe_customer_id"` // Stripe Customer ID for billing
	Name             string         `json:"name"`
	Email            string         `json:"email" gorm:"unique"`
	Password         string         `json:"password"`
	Role             string         `json:"role"` // Admin or User
	MagicLinkToken   string         `json:"magic_link_token"`
	MagicLinkExpiry  time.Time      `json:"magic_link_expiry"`
	Organizations    []Organization `gorm:"many2many:organization_users;"`
	Subscriptions    []Subscription `json:"subscriptions"`
	Active           bool           `json:"active" gorm:"default:true"` // Active status

}
type OrganizationUser struct {
	gorm.Model
	UserID               uint   `json:"user_id"`
	OrganizationID       uint   `json:"organization_id"`
	StripeSubscriptionID string `json:"stripe_subscription_id"`
}
type Product struct {
	gorm.Model
	Name        string `json:"name"`
	PriceID     string `json:"price_id"`     // Stripe Price ID
	PriceAmount int64  `json:"price_amount"` // Price per user in cents
}
type Subscription struct {
	gorm.Model
	UserID               *uint `json:"user_id"`
	OrganizationID       *uint
	PriceId              string
	StripeSubscriptionID string `json:"stripe_subscription_id"`
	Quantity             int    `json:"quantity"`                             // Number of users/seats
	Active               *bool  `gorm:"not null;default:false" json:"active"` // Subscription active status
	SubscriptionStatus   string
	ProductID            uint `json:"product_id" binding:"required"`
	UsageLimit           int  `json:"usage_limit" gorm:"default:0"` // Example usage limit

}

type UserInvite struct {
	gorm.Model
	Email                string `json:"email"`
	OrganizationID       uint   `json:"organization_id"`
	InviteToken          string `json:"invite_token"`
	IsAccepted           bool   `json:"is_accepted"`
	StripeSubscriptionID string `json:"stripe_subscription_id" binding:"required"`
}
