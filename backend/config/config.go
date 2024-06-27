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
	dsn := "host=" + os.Getenv("DB_HOST") +
		" user=" + os.Getenv("DB_USER") +
		" password=" + os.Getenv("DB_PASSWORD") +
		" dbname=" + os.Getenv("DB_NAME") +
		" port=" + os.Getenv("DB_PORT") +
		" sslmode=disable TimeZone=Asia/Shanghai"
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
	if os.Getenv("STRIPE_SECRET_KEY") == "" {
		log.Panic("STRIPE_SECRET_KEY required")
	}
}
