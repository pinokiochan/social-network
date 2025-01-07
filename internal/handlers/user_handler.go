package handlers

import (
	"encoding/json"
	"net/http"
	"log"
	"fmt"
	"github.com/pinokiochan/social-network/internal/models"
	"github.com/pinokiochan/social-network/internal/auth"
	"github.com/pinokiochan/social-network/internal/utils"
	"database/sql"
)

type UserHandler struct {
	db *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if !utils.IsValidEmail(input.Email) || !utils.IsAlpha(input.Username) {
		http.Error(w, "Invalid input format", http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	var userID int
	err = h.db.QueryRow(
		"INSERT INTO users (username, email, password, is_admin) VALUES ($1, $2, $3, $4) RETURNING id",
		input.Username, input.Email, hashedPassword, false,
	).Scan(&userID)

	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateToken(userID, false)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"user": map[string]interface{}{
			"id":       userID,
			"username": input.Username,
			"email":    input.Email,
		},
		"token": token,
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	var user models.User
	err := h.db.QueryRow("SELECT id, password, is_admin FROM users WHERE email = $1", credentials.Email).Scan(&user.ID, &user.Password, &user.IsAdmin)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := auth.CheckPasswordHash(credentials.Password, user.Password); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(user.ID, user.IsAdmin)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"token":   token,
		"user_id": user.ID,
		"is_admin": user.IsAdmin,
	})
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT id, username, email, is_admin FROM users")
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.IsAdmin); err != nil {
			http.Error(w, "Error scanning user", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
func (h *UserHandler) SendEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	to := r.Form.Get("to")
	subject := r.Form.Get("subject")
	body := r.Form.Get("body")

	if to == "" || subject == "" || body == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Use the email utility function
	err = utils.SendEmail(to, subject, body)  // Здесь вызываем функцию из utils
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		http.Error(w, fmt.Sprintf("Failed to send email: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Email sent successfully",
	})
}

