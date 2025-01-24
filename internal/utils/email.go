package utils

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/pinokiochan/social-network/internal/logger"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
	"os"
)

func SendEmail(to, subject, body, attachmentPath string) error {
	// Загрузка переменных окружения из .env файла
	err := godotenv.Load()
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to load .env file")
		return fmt.Errorf("Error loading .env file: %v", err)
	}

	// Извлечение SMTP настроек из окружения
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

	// Проверка наличия всех обязательных переменных окружения
	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" {
		logger.Log.WithFields(logrus.Fields{
			"host": smtpHost != "",
			"port": smtpPort != "",
			"user": smtpUser != "",
		}).Error("Missing SMTP configuration")
		return fmt.Errorf("SMTP configuration is missing in environment variables")
	}

	// Создание нового письма
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", smtpUser)
	mailer.SetHeader("To", to)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/plain", body)

	// Прикрепление файла, если путь указан
	if attachmentPath != "" {
		mailer.Attach(attachmentPath)
	}

	// Логирование попытки отправки письма
	logger.Log.WithFields(logrus.Fields{
		"to":      to,
		"subject": subject,
		"from":    smtpUser,
	}).Debug("Attempting to send email")

	// Отправка письма
	dialer := gomail.NewDialer(smtpHost, 587, smtpUser, smtpPass)
	err = dialer.DialAndSend(mailer)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
			"to":    to,
		}).Error("Failed to send email")
		return fmt.Errorf("failed to send email: %v", err)
	}

	// Логирование успешной отправки письма
	logger.Log.WithFields(logrus.Fields{
		"to":      to,
		"subject": subject,
	}).Info("Email sent successfully")

	return nil
}
