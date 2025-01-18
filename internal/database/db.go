package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/pinokiochan/social-network/internal/logger"
)

func ConnectToDB() (*sql.DB, error) {
	connStr := os.Getenv("DATABASE_URL") // Получаем строку подключения
	if connStr == "" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}

	logger.InfoLogger("Attempting database connection", logger.Fields{
		"connection_string": connStr,
	})

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.ErrorLogger(err, logger.Fields{
			"error": "Failed to open database connection",
		})
		return nil, fmt.Errorf("failed to open database connection: %v", err)
	}

	// Проверяем подключение
	err = db.Ping()
	if err != nil {
		logger.ErrorLogger(err, logger.Fields{
			"error": "Failed to ping database",
		})
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	logger.InfoLogger("Database connection established successfully", logger.Fields{
		"status": "connected",
	})

	return db, nil
}
