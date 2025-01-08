package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"github.com/pinokiochan/social-network/internal/models"
	"github.com/pinokiochan/social-network/internal/middleware"
	"github.com/pinokiochan/social-network/internal/logger"
	"github.com/sirupsen/logrus"
	"database/sql"
)

type PostHandler struct {
	db *sql.DB
}

func NewPostHandler(db *sql.DB) *PostHandler {
	return &PostHandler{db: db}
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Log.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		}).Warn("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var post models.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Invalid JSON format")
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	userID, err := middleware.GetUserIDFromToken(r)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Unauthorized access attempt")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = h.db.QueryRow(
		"INSERT INTO posts (user_id, content) VALUES ($1, $2) RETURNING id, created_at",
		userID, post.Content,
	).Scan(&post.ID, &post.CreatedAt)

	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error":  err.Error(),
			"userID": userID,
		}).Error("Failed to create post")
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}

	post.UserID = userID

	logger.Log.WithFields(logrus.Fields{
		"postID": post.ID,
		"userID": userID,
	}).Info("Post created successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func (h *PostHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	keyword := query.Get("keyword")
	userID := query.Get("user_id")
	date := query.Get("date")
	page := query.Get("page")
	pageSize := query.Get("page_size")

	if page == "" {
		page = "1"
	}
	if pageSize == "" {
		pageSize = "10"
	}

	offset, _ := strconv.Atoi(page)
	limit, _ := strconv.Atoi(pageSize)
	offset = (offset - 1) * limit

	baseQuery := `
		SELECT posts.id, posts.user_id, posts.content, posts.created_at, users.username 
		FROM posts 
		JOIN users ON posts.user_id = users.id
	`
	whereClause := []string{}
	args := []interface{}{}

	if keyword != "" {
		whereClause = append(whereClause, "posts.content ILIKE $"+strconv.Itoa(len(args)+1))
		args = append(args, "%"+keyword+"%")
	}
	if userID != "" {
		whereClause = append(whereClause, "posts.user_id = $"+strconv.Itoa(len(args)+1))
		args = append(args, userID)
	}
	if date != "" {
		whereClause = append(whereClause, "DATE(posts.created_at) = $"+strconv.Itoa(len(args)+1))
		args = append(args, date)
	}

	if len(whereClause) > 0 {
		baseQuery += " WHERE " + strings.Join(whereClause, " AND ")
	}

	baseQuery += " ORDER BY posts.created_at DESC LIMIT $" + strconv.Itoa(len(args)+1) + 
		" OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, limit, offset)

	logger.Log.WithFields(logrus.Fields{
		"keyword":    keyword,
		"user_id":    userID,
		"date":       date,
		"page":       page,
		"page_size":  pageSize,
	}).Debug("Fetching posts with filters")

	rows, err := h.db.Query(baseQuery, args...)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to fetch posts")
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.ID, &post.UserID, &post.Content, &post.CreatedAt, &post.Username)
		if err != nil {
			logger.Log.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("Error scanning post")
			http.Error(w, "Error scanning post", http.StatusInternalServerError)
			return
		}
		posts = append(posts, post)
	}

	logger.Log.WithFields(logrus.Fields{
		"count": len(posts),
	}).Info("Posts fetched successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		logger.Log.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		}).Warn("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var post models.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Invalid input")
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	userID, err := middleware.GetUserIDFromToken(r)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Unauthorized access attempt")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	result, err := h.db.Exec(
		"UPDATE posts SET content = $1 WHERE id = $2 AND user_id = $3",
		post.Content, post.ID, userID,
	)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error":  err.Error(),
			"postID": post.ID,
			"userID": userID,
		}).Error("Failed to update post")
		http.Error(w, "Error updating post", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		logger.Log.WithFields(logrus.Fields{
			"postID": post.ID,
			"userID": userID,
		}).Warn("Post not found or unauthorized modification attempt")
		http.Error(w, "Post not found or you don't have permission to edit it", http.StatusForbidden)
		return
	}

	logger.Log.WithFields(logrus.Fields{
		"postID": post.ID,
		"userID": userID,
	}).Info("Post updated successfully")

	w.WriteHeader(http.StatusOK)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		logger.Log.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		}).Warn("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var post models.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Invalid input")
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	userID, err := middleware.GetUserIDFromToken(r)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Unauthorized access attempt")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	result, err := h.db.Exec("DELETE FROM posts WHERE id = $1 AND user_id = $2", post.ID, userID)
	if err != nil {
		logger.Log.WithFields(logrus.Fields{
			"error":  err.Error(),
			"postID": post.ID,
			"userID": userID,
		}).Error("Failed to delete post")
		http.Error(w, "Error deleting post", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		logger.Log.WithFields(logrus.Fields{
			"postID": post.ID,
			"userID": userID,
		}).Warn("Post not found or unauthorized deletion attempt")
		http.Error(w, "Post not found or you don't have permission to delete it", http.StatusForbidden)
		return
	}

	logger.Log.WithFields(logrus.Fields{
		"postID": post.ID,
		"userID": userID,
	}).Info("Post deleted successfully")

	w.WriteHeader(http.StatusOK)
}

