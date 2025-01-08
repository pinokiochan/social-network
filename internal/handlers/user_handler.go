package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pinokiochan/social-network/internal/models"
	"github.com/pinokiochan/social-network/internal/auth"
	"github.com/pinokiochan/social-network/internal/utils"
	"github.com/pinokiochan/social-network/internal/logger"
	"database/sql"
	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	db *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Log.WithFields(logrus.Fields{
			"path":   r.URL.Path,
			"method": r.Method,
		}).Error(fmt.Errorf("method not allowed: %s", r.Method))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": "Invalid JSON format",
			"path":  r.URL.Path,
		}).Error(err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if !utils.IsValidEmail(input.Email) || !utils.IsAlpha(input.Username) {
		logger.Log.WithFields(logrus.Fields{
			"email":    input.Email,
			"username": input.Username,
		}).Error(fmt.Errorf("invalid input format"))
		http.Error(w, "Invalid input format", http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": "Error processing password",
		}).Error(err)
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	var userID int
	err = h.db.QueryRow(
		"INSERT INTO users (username, email, password, is_admin) VALUES ($1, $2, $3, $4) RETURNING id",
		input.Username, input.Email, hashedPassword, false,
	).Scan(&userID)

	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": "Error creating user",
		}).Error(err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateToken(userID, false)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": "Error generating token",
			"userID": userID,
		}).Error(err)
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	logger.Log.WithFields(logrus.Fields{
		"userID":   userID,
		"username": input.Username,
		"email":    input.Email,
	}).Info("User registered successfully")

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
		logger.Log.WithFields(logrus.Fields{
			"path":   r.URL.Path,
			"method": r.Method,
		}).Error(fmt.Errorf("method not allowed: %s", r.Method))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": "Invalid JSON format",
			"path":  r.URL.Path,
		}).Error(err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	var user models.User
	err := h.db.QueryRow("SELECT id, password, is_admin FROM users WHERE email = $1", credentials.Email).Scan(&user.ID, &user.Password, &user.IsAdmin)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": "Invalid credentials",
			"email": credentials.Email,
		}).Error(err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := auth.CheckPasswordHash(credentials.Password, user.Password); err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": "Invalid credentials",
			"email": credentials.Email,
		}).Error(err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(user.ID, user.IsAdmin)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": "Error generating token",
			"userID": user.ID,
		}).Error(err)
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	logger.Log.WithFields(logrus.Fields{
		"userID":  user.ID,
		"email":   credentials.Email,
		"isAdmin": user.IsAdmin,
	}).Info("User logged in successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"token":    token,
		"user_id":  user.ID,
		"is_admin": user.IsAdmin,
	})
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT id, username, email, is_admin FROM users")
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": "Error fetching users",
		}).Error(err)
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.IsAdmin); err != nil {
			logger.Log.WithFields(logrus.Fields{
				"error": "Error scanning user",
			}).Error(err)
			http.Error(w, "Error scanning user", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	logger.Log.WithFields(logrus.Fields{
		"count": len(users),
	}).Info("Users fetched successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}


