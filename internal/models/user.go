package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        int       `json:"id"`        // Unique identifier for the user
	Username  string    `json:"username"`  // The user's username
	Email     string    `json:"email"`     // The user's email address
	Password  string    `json:"-"`         // The user's password (not exposed in JSON response)
	IsAdmin   bool      `json:"is_admin"`  // Flag to determine if the user is an admin
	CreatedAt time.Time `json:"created_at"`// The timestamp when the user was created
	UpdatedAt time.Time `json:"updated_at"`// The timestamp when the user was last updated
}

// NewUser creates and returns a new user instance with the provided username, email, and password
func NewUser(username, email, password string) *User {
	now := time.Now()
	return &User{
		Username:  username,
		Email:     email,
		Password:  password,
		IsAdmin:   false, // Default: new users are not admins
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// SetPassword updates the user's password and sets the updated timestamp
func (u *User) SetPassword(password string) {
	u.Password = password
	u.UpdatedAt = time.Now()
}

// SetAdmin updates the user's admin status and sets the updated timestamp
func (u *User) SetAdmin(isAdmin bool) {
	u.IsAdmin = isAdmin
	u.UpdatedAt = time.Now()
}
