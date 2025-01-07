package utils

import (
	"fmt"
	"net/smtp"
	"os"
	"github.com/joho/godotenv"
	
	
)

// SendEmail sends an email to a specified address with the given subject and body
func SendEmail(to, subject, body string) error {
	// Load .env file to read environment variables
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("Error loading .env file: %v", err)
	}

	// Get SMTP server configuration from environment variables
	smtpHost := os.Getenv("SMTP_HOST") // e.g. smtp.mail.me.com
	smtpPort := os.Getenv("SMTP_PORT") // e.g. 587
	smtpUser := os.Getenv("SMTP_USER") // e.g. pinokiochan_n@icloud.com
	smtpPass := os.Getenv("SMTP_PASS") // e.g. password or app-specific password

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("SMTP configuration is missing in environment variables")
	}

	from := smtpUser
	recipients := []string{to}

	// Compose the message with proper headers
	message := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		from, to, subject, body,
	))

	// Connect to the SMTP server with authentication
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	address := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	err = smtp.SendMail(address, auth, from, recipients, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}