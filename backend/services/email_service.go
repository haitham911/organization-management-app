package services

import (
	"log"

	"github.com/resend/resend-go/v2"
)

func SendEmail(toEmail, subject, plainTextContent, htmlContent string) error {

	apiKey := "re_G6Epm53o_8kbG4sjrpTJRkMRHDzEAbG5W"

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
