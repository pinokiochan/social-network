package utils

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/pinokiochan/social-network/internal/logger"
	"github.com/sirupsen/logrus"
	"net/smtp"
	"os"
)

func SendEmail(to, subject, body string) error {
	err := godotenv.Load()
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to load .env file")
		return fmt.Errorf("Error loading .env file: %v", err)
	}

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" {
		logger.Log.WithFields(logrus.Fields{
			"host": smtpHost != "",
			"port": smtpPort != "",
			"user": smtpUser != "",
		}).Error("Missing SMTP configuration")
		return fmt.Errorf("SMTP configuration is missing in environment variables")
	}

	from := smtpUser
	recipients := []string{to}

	message := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		from, to, subject, body,
	))

	logger.Log.WithFields(logrus.Fields{
		"to":      to,
		"subject": subject,
		"from":    from,
	}).Debug("Attempting to send email")

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	address := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	err = smtp.SendMail(address, auth, from, recipients, message)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
			"to":    to,
		}).Error("Failed to send email")
		return fmt.Errorf("failed to send email: %v", err)
	}

	logger.Log.WithFields(logrus.Fields{
		"to":      to,
		"subject": subject,
	}).Info("Email sent successfully")

	return nil
}
