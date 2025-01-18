package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pinokiochan/social-network/internal/logger"
)

func ConnectToDB() (*sql.DB, error) {
	connStr := "host=127.0.0.1 port=5432 user=postgres password=0000 dbname=social-network sslmode=disable"
	
	logger.InfoLogger("Attempting database connection", logger.Fields{
		"host": "127.0.0.1",
		"port": 5432,
		"db":   "social-network",
	})

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.ErrorLogger(err, logger.Fields{
			"error": "Failed to open database connection",
		})
		return nil, fmt.Errorf("failed to open database connection: %v", err)
	}

	// Test the connection
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