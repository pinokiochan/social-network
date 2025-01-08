package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"os"
	"io"
	"path/filepath"
	"github.com/pinokiochan/social-network/internal/models"
	"github.com/pinokiochan/social-network/internal/utils"
	"github.com/pinokiochan/social-network/internal/logger"
	"github.com/sirupsen/logrus"
)

type AdminHandler struct {
	db *sql.DB
	wg *sync.WaitGroup
}

type AdminStats struct {
	TotalUsers     int `json:"total_users"`
	TotalPosts     int `json:"total_posts"`
	TotalComments  int `json:"total_comments"`
	ActiveUsers24h int `json:"active_users_24h"`
}

func NewAdminHandler(db *sql.DB, wg *sync.WaitGroup) *AdminHandler {
	return &AdminHandler{db: db, wg: wg}
}

func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	var stats AdminStats

	err := h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to get user stats")
		http.Error(w, "Error getting user stats", http.StatusInternalServerError)
		return
	}

	err = h.db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&stats.TotalPosts)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to get post stats")
		http.Error(w, "Error getting post stats", http.StatusInternalServerError)
		return
	}

	err = h.db.QueryRow("SELECT COUNT(*) FROM comments").Scan(&stats.TotalComments)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to get comment stats")
		http.Error(w, "Error getting comment stats", http.StatusInternalServerError)
		return
	}

	err = h.db.QueryRow(`
		SELECT COUNT(DISTINCT user_id) 
		FROM (
			SELECT user_id FROM posts WHERE created_at > NOW() - INTERVAL '24 hours'
			UNION
			SELECT user_id FROM comments WHERE created_at > NOW() - INTERVAL '24 hours'
		) as active_users
	`).Scan(&stats.ActiveUsers24h)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to get active users stats")
		http.Error(w, "Error getting active users stats", http.StatusInternalServerError)
		return
	}

	logger.Log.WithFields(logrus.Fields{
		"total_users":      stats.TotalUsers,
		"total_posts":      stats.TotalPosts,
		"total_comments":   stats.TotalComments,
		"active_users_24h": stats.ActiveUsers24h,
	}).Info("Admin stats retrieved successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *AdminHandler) BroadcastEmailToSelectedUsers(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		logger.Log.WithError(err).Error("Failed to parse multipart form")
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	subject := r.FormValue("subject")
	body := r.FormValue("body")
	users := r.Form["users[]"]

	file, header, err := r.FormFile("attachment")
	if err != nil && err != http.ErrMissingFile {
		logger.Log.WithError(err).Error("Error retrieving file from form")
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}

	var attachmentPath string
	if file != nil {
		defer file.Close()
		attachmentPath = filepath.Join(os.TempDir(), header.Filename)
		outFile, err := os.Create(attachmentPath)
		if err != nil {
			logger.Log.WithError(err).Error("Failed to create temporary file")
			http.Error(w, "Failed to process attachment", http.StatusInternalServerError)
			return
		}
		defer outFile.Close()
		if _, err := io.Copy(outFile, file); err != nil {
			logger.Log.WithError(err).Error("Failed to save attachment")
			http.Error(w, "Failed to process attachment", http.StatusInternalServerError)
			return
		}
	}

	logger.Log.WithFields(logrus.Fields{
		"subject":     subject,
		"user_count":  len(users),
		"has_attachment": attachmentPath != "",
	}).Info("Starting email broadcast")

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		for _, userEmail := range users {
			err := utils.SendEmail(userEmail, subject, body, attachmentPath)
			if err != nil {
				logger.Log.WithError(err).WithField("email", userEmail).Error("Failed to send broadcast email")
				continue
			}
			logger.Log.WithField("email", userEmail).Info("Broadcast email sent successfully")
		}
		if attachmentPath != "" {
			os.Remove(attachmentPath)
		}
		logger.Log.Info("Email broadcast completed")
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Emails are being sent",
	})
}

func (h *AdminHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT id, username, email, is_admin FROM users")
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to fetch users")
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.IsAdmin); err != nil {
			logger.Log.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("Error scanning user")
			http.Error(w, "Error scanning user", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	logger.Log.WithFields(logrus.Fields{
		"user_count": len(users),
	}).Info("Users fetched successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	if userID == "" {
		logger.Log.Warn("Missing user ID in delete request")
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(userID)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    userID,
		}).Error("Invalid user ID format")
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
			"id":    id,
		}).Error("Failed to delete user")
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		logger.Log.WithFields(logrus.Fields{
			"id": id,
		}).Warn("User not found for deletion")
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	logger.Log.WithFields(logrus.Fields{
		"id": id,
	}).Info("User deleted successfully")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
}

func (h *AdminHandler) EditUser(w http.ResponseWriter, r *http.Request) {
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
		Email    string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to decode payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if payload.ID == 0 || payload.Username == "" || payload.Email == "" {
		logger.Log.WithFields(logrus.Fields{
			"id":       payload.ID,
			"username": payload.Username,
			"email":    payload.Email,
		}).Warn("Missing required fields")
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec(`
		UPDATE users 
		SET username = $1, email = $2, updated_at = NOW()
		WHERE id = $3
	`, payload.Username, payload.Email, payload.ID)
	
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
		"email":    payload.Email,
	}).Info("User updated successfully")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User updated successfully"})
}

