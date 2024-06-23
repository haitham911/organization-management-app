package services

import (
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendEmail(toEmail string, link string) error {
	from := mail.NewEmail("Your App Name", "your-email@example.com")
	subject := "You're Invited!"
	to := mail.NewEmail("User", toEmail)
	plainTextContent := "Click the following link to join: " + link
	htmlContent := "<p>Click the following link to join: <a href=\"" + link + "\">Join Now</a></p>"
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	_, err := client.Send(message)
	return err
}
