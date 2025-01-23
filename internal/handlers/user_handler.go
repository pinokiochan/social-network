package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/pinokiochan/social-network/internal/auth"
	"github.com/pinokiochan/social-network/internal/logger"
	"github.com/pinokiochan/social-network/internal/models"
	"github.com/pinokiochan/social-network/internal/utils"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
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

	// generate 6 digit code
	//send email

	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": "Error processing password",
		}).Error(err)
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	var userID int
	query := "INSERT INTO users (username, email, password, is_admin, is_active) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	_, err = h.db.Exec(query, input.Username, input.Email, hashedPassword, false, false)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": "Error inserting user",
		}).Error(err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	code := utils.GenerateCode()
	query = "INSERT INTO inactive_users (email, code) VALUES ($1, $2)"
	_, err = h.db.Exec(query, input.Email, code)

	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": "Error creating user",
		}).Error(err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// token, err := auth.GenerateToken(userID, false)
	// if err != nil {
	// 	logger.Log.WithFields(logrus.Fields{
	// 		"error": "Error generating token",
	// 		"userID": userID,
	// 	}).Error(err)
	// 	http.Error(w, "Error generating token", http.StatusInternalServerError)
	// 	return
	// }

	logger.Log.WithFields(logrus.Fields{
		"userID":   userID,
		"username": input.Username,
		"email":    input.Email,
	}).Info("User registered successfully")

	err = utils.SendEmail(input.Email, "Verification Code", fmt.Sprintf("Verify your email via this 4-digit code: %v", code), "")
	if err != nil {
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
		// "token": token,
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

	getStatus := "SELECT is_active FROM users WHERE email = $1"

	// Variable to hold the is_active value
	var isActive bool

	// Use QueryRow to get the result
	err = h.db.QueryRow(getStatus, credentials.Email).Scan(&isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return
		} else {
			return
		}
	}

	// Check the is_active value
	if !isActive {
		http.Error(w, "Email not verified", 400)
		return
	}

	token, err := auth.GenerateToken(user.ID, user.IsAdmin)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error":  "Error generating token",
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

func (h *UserHandler) Verify(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email string `json:"email"`
		Code  int    `json:"code"`
	}

	json.NewDecoder(r.Body).Decode(&credentials)

	row := h.db.QueryRow("SELECT inactive_users_id FROM inactive_users WHERE email = $1 AND code = $2", credentials.Email, credentials.Code)
	var userID int
	err := row.Scan(&userID)
	if err == sql.ErrNoRows {
		logger.Log.WithFields(logrus.Fields{
			"error": "Invalid verification code",
			"email": credentials.Email,
			"code":  credentials.Code,
		}).Error(fmt.Errorf("invalid verification code"))
		http.Error(w, "Invalid verification code", http.StatusNotFound)
		return
	} else if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": "Error fetching user ID",
			"email": credentials.Email,
			"code":  credentials.Code,
		}).Error(err)
		http.Error(w, "Error fetching user ID", http.StatusInternalServerError)
		return
	}

	query := "UPDATE users set is_active=true where email=$1"
	_, err = h.db.Exec(query, credentials.Email)

	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error":  "Error activating user",
			"userID": userID,
		}).Error(err)
		http.Error(w, "Error activating user", http.StatusInternalServerError)
		return
	}

	logger.Log.WithFields(logrus.Fields{
		"userID": userID,
		"email":  credentials.Email,
	}).Info("User verified successfully")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
	})
}

// user-profile
// logic for update current user data: password, username
// user activity
// get posts from current user
// get comments from current user
func (h *UserHandler) UserUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Log.WithFields(logrus.Fields{
			"method": r.Method,
		}).Warn("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to decode payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if payload.ID == 0 || payload.Username == "" || payload.Password == "" {
		logger.Log.WithFields(logrus.Fields{
			"id":       payload.ID,
			"username": payload.Username,
			"password": payload.Password,
		}).Warn("Missing required fields")
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Hash the password before saving it
	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Error hashing password")
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Update user data in database
	result, err := h.db.Exec(`
		UPDATE users 
		SET username = $1, password = $2, updated_at = NOW()
		WHERE id = $3
	`, payload.Username, hashedPassword, payload.ID)

	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    payload.ID,
		}).Error("Failed to update user")
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		logger.Log.WithFields(logrus.Fields{
			"id": payload.ID,
		}).Warn("User not found for update")
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	logger.Log.WithFields(logrus.Fields{
		"id":       payload.ID,
		"username": payload.Username,
	}).Info("User updated successfully")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User updated successfully"})
}
func (h *UserHandler) UserData(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Обработка получения данных пользователя
		userID := r.URL.Query().Get("id")
		if userID == "" {
			logger.Log.Warn("User ID is required")
			http.Error(w, `{"error": "User ID is required"}`, http.StatusBadRequest)
			return
		}

		var user struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			IsAdmin  bool   `json:"is_admin"`
		}

		err := h.db.QueryRow(`
			SELECT username, email, is_admin
			FROM users 
			WHERE id = $1
		`, userID).Scan(&user.Username, &user.Email, &user.IsAdmin)
		if err != nil {
			if err == sql.ErrNoRows {
				// Пользователь не найден, возвращаем ошибку в JSON формате
				logger.Log.WithFields(logrus.Fields{
					"id": userID,
				}).Warn("User not found")
				http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
			} else {
				// Ошибка при запросе, возвращаем внутреннюю ошибку в JSON формате
				logger.Log.WithFields(logrus.Fields{
					"error": err.Error(),
					"id":    userID,
				}).Error("Failed to fetch user data")
				http.Error(w, `{"error": "Failed to fetch user data"}`, http.StatusInternalServerError)
			}
			return
		}

		// Логируем успешное получение данных пользователя
		logger.Log.WithFields(logrus.Fields{
			"id":       userID,
			"username": user.Username,
			"email":    user.Email,
		}).Info("User data retrieved successfully")

		// Возвращаем данные пользователя в JSON формате
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
		return
	}
}

// UserPosts handles the request to get posts for a specific user
func (h *UserHandler) UserPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		logger.Log.WithFields(logrus.Fields{
			"method": r.Method,
		}).Warn("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from query parameters
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		logger.Log.Warn("User ID is required")
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Warn("Invalid user ID")
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Fetch posts from the database
	rows, err := h.db.Query(`
		SELECT content, created_at
		FROM posts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error":  err.Error(),
			"userID": userID,
		}).Error("Failed to fetch user posts")
		http.Error(w, "Failed to fetch user posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []struct {
		Content   string `json:"content"`
		CreatedAt string `json:"created_at"`
	}

	for rows.Next() {
		var post struct {
			Content   string `json:"content"`
			CreatedAt string `json:"created_at"`
		}
		if err := rows.Scan(&post.Content, &post.CreatedAt); err != nil {
			logger.Log.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("Error scanning post row")
			continue
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Error iterating over post rows")
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}

	// Log successful retrieval of posts
	logger.Log.WithFields(logrus.Fields{
		"userID": userID,
		"count":  len(posts),
	}).Info("User posts retrieved successfully")

	// Return posts as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}
