package utils

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/pinokiochan/social-network/internal/logger"
	"github.com/sirupsen/logrus"
	"net/smtp"
	"os"
	"github.com/jordan-wright/email"
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
	e := email.NewEmail()
	e.From = smtpUser
	e.To = []string{to}
	e.Subject = subject
	e.Text = []byte(body)

	// Прикрепление файла, если путь указан
	if attachmentPath != "" {
		_, err := e.AttachFile(attachmentPath)
		if err != nil {
			logger.Log.WithFields(logrus.Fields{
				"error": err.Error(),
				"path":  attachmentPath,
			}).Error("Failed to attach file to email")
			return fmt.Errorf("failed to attach file: %v", err)
		}
	}

	// Логирование попытки отправки письма
	logger.Log.WithFields(logrus.Fields{
		"to":      to,
		"subject": subject,
		"from":    smtpUser,
	}).Debug("Attempting to send email")

	// Отправка письма
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	address := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	err = e.Send(address, auth)
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
