package models

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type Organization struct {
	ID               uint           `gorm:"primarykey"`
	CreatedAt        time.Time      `json:"created_at,omitempty"`
	UpdatedAt        time.Time      `json:"updated_at,omitempty"`
	Name             string         `json:"name"`
	Email            string         `json:"email" gorm:"unique"`
	StripeCustomerID string         `json:"stripe_customer_id"` // Stripe Customer ID for billing
	Users            []User         `gorm:"many2many:organization_users;" json:"users"`
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
	Password         string         `json:"-"`
	MagicLinkToken   string         `json:"magic_link_token"`
	MagicLinkExpiry  time.Time      `json:"magic_link_expiry"`
	Organizations    []Organization `gorm:"many2many:organization_users;"`
	Subscriptions    []Subscription `json:"subscriptions"`
	Active           *bool          `json:"active" gorm:"default:false"` // Active status

}

type OrganizationUser struct {
	ID                   uint `gorm:"primarykey"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
	UserID               uint         `json:"user_id"`
	OrganizationID       uint         `json:"organization_id"`
	StripeSubscriptionID string       `json:"stripe_subscription_id"`
	Role                 string       `json:"role"` // Role in the organization
	Organization         Organization `json:"organization"`
	User                 User         `json:"user"`
}
type Product struct {
	gorm.Model
	Name        string `json:"name"`
	PriceID     string `json:"price_id"`     // Stripe Price ID
	PriceAmount int64  `json:"price_amount"` // Price per user in cents
}
type Subscription struct {
	ID                   uint      `gorm:"primarykey" json:"ID,omitempty"`
	CreatedAt            time.Time `json:"created_at,omitempty"`
	UpdatedAt            time.Time `json:"updated_at,omitempty"`
	UserID               *uint     `json:"user_id,omitempty"`
	OrganizationID       *uint     `json:"organization_id,omitempty"`
	PriceId              string    `json:"price_id,omitempty"`
	StripeSubscriptionID string    `json:"stripe_subscription_id,omitempty"`
	Quantity             int       `json:"quantity,omitempty"`                             // Number of users/seats
	Active               *bool     `gorm:"not null;default:false" json:"active,omitempty"` // Subscription active status
	SubscriptionStatus   string    `json:"subscription_status,omitempty"`
	ProductID            string      `json:"product_id,omitempty" binding:"required"`
	UsageLimit           int       `json:"usage_limit,omitempty" gorm:"default:0"` // Example usage limit

}

type UserInvite struct {
	ID                   uint `gorm:"primarykey"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
	Email                string `json:"email"`
	OrganizationID       uint   `json:"organization_id"`
	InviteToken          string `json:"invite_token"`
	IsAccepted           bool   `json:"is_accepted"`
	StripeSubscriptionID string `json:"stripe_subscription_id" binding:"required"`
	Role                 string `json:"role"`
}
type UserWithRoles struct {
	Role                    string       `json:"role"`
	Organization            Organization `json:"organization"`
	OrgStripeSubscriptionID string       `json:"org_subscription_id"`
}
