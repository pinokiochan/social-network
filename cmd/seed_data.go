package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/pinokiochan/social-network/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func seed_data() {
	// Connect to the database, and handle the error
	db, err := database.ConnectToDB()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	defer db.Close()

	// Seed users
	for i := 1; i <= 50; i++ {
		username := fmt.Sprintf("user%d", i)
		email := fmt.Sprintf("%d@astanait.edu.kz", 100000+i)
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		_, err := db.Exec("INSERT INTO users (username, email, password) VALUES ($1, $2, $3)", username, email, string(hashedPassword))
		if err != nil {
			log.Printf("Error inserting user: %v", err)
		}
	}

	// Create admin user
	adminPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	_, err = db.Exec(
		"INSERT INTO users (username, email, password, is_admin) VALUES ($1, $2, $3, $4)",
		"admin",
		"admin@example.com",
		string(adminPassword),
		true,
	)
	if err != nil {
		log.Printf("Error creating admin user: %v", err)
	}

	// Seed posts
	for i := 1; i <= 100; i++ {
		userID := rand.Intn(50) + 1
		content := fmt.Sprintf("This is a sample post number %d. It contains some interesting content that a user might write.", i)
		createdAt := time.Now().Add(-time.Duration(rand.Intn(30*24)) * time.Hour)
		_, err := db.Exec("INSERT INTO posts (user_id, content, created_at) VALUES ($1, $2, $3)", userID, content, createdAt)
		if err != nil {
			log.Printf("Error inserting post: %v", err)
		}
	}

	// Seed comments
	for i := 1; i <= 200; i++ {
		userID := rand.Intn(50) + 1
		postID := rand.Intn(100) + 1
		content := fmt.Sprintf("This is a sample comment number %d. It's a response to the post.", i)
		createdAt := time.Now().Add(-time.Duration(rand.Intn(30*24)) * time.Hour)
		_, err := db.Exec("INSERT INTO comments (user_id, post_id, content, created_at) VALUES ($1, $2, $3, $4)", userID, postID, content, createdAt)
		if err != nil {
			log.Printf("Error inserting comment: %v", err)
		}
	}

	fmt.Println("Seed data inserted successfully")
}
