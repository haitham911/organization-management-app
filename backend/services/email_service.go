package services

import (
	"errors"
	"log"
	"os"

	"github.com/resend/resend-go/v2"
)

func SendEmail(toEmail, subject, plainTextContent, htmlContent string) error {

	apiKey := os.Getenv("EMAIL_API_KEY")
	if apiKey == "" {
		return errors.New("EMAIL_API_KEY required")

	}
	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    "onboarding@resend.dev",
		To:      []string{toEmail},
		Subject: subject,
		Html:    htmlContent,
		Text:    plainTextContent,
	}

	_, err := client.Emails.Send(params)
	if err != nil {
		log.Println(err)
	}

	return err
}
