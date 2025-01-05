package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func ConnectToDB() *sql.DB {
	connStr := "host=127.0.0.1 port=5432 user=postgres password=0000 dbname=social-network sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	return db
}
