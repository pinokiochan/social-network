package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"log"
	"strings"
	"github.com/pinokiochan/social-network/internal/models"
	"github.com/pinokiochan/social-network/internal/utils"
)

type AdminHandler struct {
	db *sql.DB
}

type AdminStats struct {
	TotalUsers     int `json:"total_users"`
	TotalPosts     int `json:"total_posts"`
	TotalComments  int `json:"total_comments"`
	ActiveUsers24h int `json:"active_users_24h"`
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

// GetUserByID retrieves a user by ID from the database
func (h *AdminHandler) GetUserByID(id int) (*models.User, error) {
	query := `SELECT id, username, email, is_admin, created_at, updated_at FROM users WHERE id = $1`
	user := models.User{}
	err := h.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Email, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates user details in the database
func (h *AdminHandler) UpdateUser(user *models.User) error {
	query := `UPDATE users SET username = $1, email = $2, is_admin = $3, updated_at = NOW() WHERE id = $4`
	result, err := h.db.Exec(query, user.Username, user.Email, user.IsAdmin, user.ID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no user updated, possibly invalid ID")
	}
	return nil
}

func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	var stats AdminStats

	err := h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	if err != nil {
		http.Error(w, "Error getting user stats", http.StatusInternalServerError)
		return
	}

	err = h.db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&stats.TotalPosts)
	if err != nil {
		http.Error(w, "Error getting post stats", http.StatusInternalServerError)
		return
	}

	err = h.db.QueryRow("SELECT COUNT(*) FROM comments").Scan(&stats.TotalComments)
	if err != nil {
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
		http.Error(w, "Error getting active users stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *AdminHandler) BroadcastEmailToSelectedUsers(w http.ResponseWriter, r *http.Request) {
    var request struct {
        Subject string   `json:"subject"`
        Body    string   `json:"body"`
        Users   []string `json:"users"` // A list of user emails
    }

    // Decode the request body
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Basic validation
    if len(request.Users) == 0 || strings.TrimSpace(request.Subject) == "" || strings.TrimSpace(request.Body) == "" {
        http.Error(w, "Please provide users, subject, and body", http.StatusBadRequest)
        return
    }

    // Use goroutines to send emails concurrently
    go func() {
        for _, userEmail := range request.Users {
            err := utils.SendEmail(userEmail, request.Subject, request.Body)
            if err != nil {
                log.Printf("Failed to send email to %s: %v", userEmail, err)
                continue
            }
            log.Printf("Email sent to %s successfully", userEmail)
        }
    }()

    // Respond immediately with a success message
    response := map[string]interface{}{
        "success": true,
        "message": "Emails are being sent",
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}


func (h *AdminHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
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

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	_, err = h.db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
}

func (h *AdminHandler) EditUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	// Parse JSON payload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("Error decoding payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Log received payload for debugging
	log.Printf("EditUser payload: %+v", payload)

	// Validate input
	if payload.ID == 0 || payload.Username == "" || payload.Email == "" {
		log.Printf("Validation error: ID=%d, Username=%s, Email=%s", payload.ID, payload.Username, payload.Email)
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Update user in database
	_, err := h.db.Exec(`
		UPDATE users 
		SET username = $1, email = $2, updated_at = NOW()
		WHERE id = $3
	`, payload.Username, payload.Email, payload.ID)
	if err != nil {
		log.Printf("Database update error: %v", err)
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User updated successfully"})
}
