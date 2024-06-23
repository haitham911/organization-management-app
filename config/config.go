package config

import (
	"log"
	"organization-management-app/models"
	"os"

	"github.com/stripe/stripe-go/v72"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() *gorm.DB {
	dsn := "host=localhost user=your_user password=your_password dbname=org_db port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(&models.Organization{}, &models.User{}, &models.Product{}, &models.Subscription{})
	DB = db
	return db
}

func InitStripe() {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
}
