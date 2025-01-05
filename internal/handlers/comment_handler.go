package handlers

import (
	"database/sql"
	"encoding/json"
	"github.com/pinokiochan/social-network/internal/middleware"
	"github.com/pinokiochan/social-network/internal/models"
	"net/http"
)

type CommentHandler struct {
	db *sql.DB
}

func NewCommentHandler(db *sql.DB) *CommentHandler {
	return &CommentHandler{db: db}
}

func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var comment models.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	userID, err := middleware.GetUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = h.db.QueryRow(
		"INSERT INTO comments (post_id, user_id, content) VALUES ($1, $2, $3) RETURNING id, created_at",
		comment.PostID, userID, comment.Content,
	).Scan(&comment.ID, &comment.CreatedAt)

	if err != nil {
		http.Error(w, "Error creating comment", http.StatusInternalServerError)
		return
	}

	comment.UserID = userID

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

func (h *CommentHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT comments.id, comments.post_id, comments.user_id, comments.content, comments.created_at, users.username FROM comments JOIN users ON comments.user_id = users.id")
	if err != nil {
		http.Error(w, "Error fetching comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt, &comment.Username)
		if err != nil {
			http.Error(w, "Error scanning comment", http.StatusInternalServerError)
			return
		}
		comments = append(comments, comment)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var comment models.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	userID, err := middleware.GetUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	result, err := h.db.Exec("UPDATE comments SET content = $1 WHERE id = $2 AND user_id = $3", comment.Content, comment.ID, userID)
	if err != nil {
		http.Error(w, "Error updating comment", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Comment not found or you don't have permission to edit it", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var comment models.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	userID, err := middleware.GetUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if the user is the author of the comment or the author of the post
	var postUserID int
	err = h.db.QueryRow("SELECT user_id FROM posts WHERE id = (SELECT post_id FROM comments WHERE id = $1)", comment.ID).Scan(&postUserID)
	if err != nil {
		http.Error(w, "Error fetching post information", http.StatusInternalServerError)
		return
	}

	var result sql.Result
	if userID == postUserID {
		// If the user is the post author, they can delete any comment
		result, err = h.db.Exec("DELETE FROM comments WHERE id = $1", comment.ID)
	} else {
		// Otherwise, users can only delete their own comments
		result, err = h.db.Exec("DELETE FROM comments WHERE id = $1 AND user_id = $2", comment.ID, userID)
	}

	if err != nil {
		http.Error(w, "Error deleting comment", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Comment not found or you don't have permission to delete it", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
}
