package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Post struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

var jwtKey = []byte("your_secret_key")

type Claims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

func connectToDB() *sql.DB {
	connStr := "host=127.0.0.1 port=5432 user=postgres password=0000 dbname=social-network sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	return db
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "frontend/index.html")
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	db := connectToDB()
	defer db.Close()

	rows, err := db.Query("SELECT id, username, email FROM users")
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Username, &user.Email)
		if err != nil {
			log.Println("Error scanning row:", err)
			continue
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating over rows:", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func getPosts(w http.ResponseWriter, r *http.Request) {
	db := connectToDB()
	defer db.Close()

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

	offset, err := strconv.Atoi(page)
	if err != nil {
		http.Error(w, "Invalid page number", http.StatusBadRequest)
		return
	}
	limit, err := strconv.Atoi(pageSize)
	if err != nil {
		http.Error(w, "Invalid page size", http.StatusBadRequest)
		return
	}
	offset = (offset - 1) * limit

	baseQuery := "SELECT posts.id, posts.user_id, posts.content, posts.created_at, users.username FROM posts JOIN users ON posts.user_id = users.id"
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

	baseQuery += " ORDER BY posts.created_at DESC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, limit, offset)

	rows, err := db.Query(baseQuery, args...)
	if err != nil {
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.UserID, &post.Content, &post.CreatedAt, &post.Username)
		if err != nil {
			log.Println("Error scanning row:", err)
			continue
		}
		posts = append(posts, post)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func createPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db := connectToDB()
	defer db.Close()

	var post Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = db.QueryRow(
		"INSERT INTO posts (user_id, content, created_at) VALUES ($1, $2, NOW()) RETURNING id, created_at",
		userID, post.Content,
	).Scan(&post.ID, &post.CreatedAt)
	if err != nil {
		log.Printf("Error creating post: %v", err)
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}

	post.UserID = userID

	// Fetch the username for the post
	row := db.QueryRow("SELECT username FROM users WHERE id = $1", userID)
	err = row.Scan(&post.Username)
	if err != nil {
		log.Printf("Error fetching username: %v", err)
		http.Error(w, "Error fetching username", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func createComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db := connectToDB()
	defer db.Close()

	var comment Comment
	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if comment.Content == "" || comment.PostID == 0 {
		http.Error(w, "Content and post_id are required", http.StatusBadRequest)
		return
	}

	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = db.QueryRow(
		"INSERT INTO comments (post_id, user_id, content, created_at) VALUES ($1, $2, $3, NOW()) RETURNING id, created_at",
		comment.PostID, userID, comment.Content,
	).Scan(&comment.ID, &comment.CreatedAt)
	if err != nil {
		log.Printf("Error creating comment: %v", err)
		http.Error(w, "Error creating comment", http.StatusInternalServerError)
		return
	}

	comment.UserID = userID

	// Fetch the username for the comment
	row := db.QueryRow("SELECT username FROM users WHERE id = $1", userID)
	err = row.Scan(&comment.Username)
	if err != nil {
		log.Printf("Error fetching username: %v", err)
		http.Error(w, "Error fetching username", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

func getComments(w http.ResponseWriter, r *http.Request) {
	db := connectToDB()
	defer db.Close()

	rows, err := db.Query("SELECT comments.id, comments.post_id, comments.user_id, comments.content, comments.created_at, users.username FROM comments JOIN users ON comments.user_id = users.id")
	if err != nil {
		http.Error(w, "Error fetching comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt, &comment.Username)
		if err != nil {
			log.Println("Error scanning row:", err)
			continue
		}
		comments = append(comments, comment)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

func updatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db := connectToDB()
	defer db.Close()

	var post Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	result, err := db.Exec("UPDATE posts SET content = $1 WHERE id = $2 AND user_id = $3", post.Content, post.ID, userID)
	if err != nil {
		http.Error(w, "Error updating post", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Post not found or you don't have permission to edit it", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db := connectToDB()
	defer db.Close()

	var post Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	result, err := db.Exec("DELETE FROM posts WHERE id = $1 AND user_id = $2", post.ID, userID)
	if err != nil {
		http.Error(w, "Error deleting post", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Post not found or you don't have permission to delete it", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func updateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db := connectToDB()
	defer db.Close()

	var comment Comment
	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	result, err := db.Exec("UPDATE comments SET content = $1 WHERE id = $2 AND user_id = $3", comment.Content, comment.ID, userID)
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

func deleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db := connectToDB()
	defer db.Close()

	var comment Comment
	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if the user is the author of the comment or the author of the post
	var postUserID int
	err = db.QueryRow("SELECT user_id FROM posts WHERE id = (SELECT post_id FROM comments WHERE id = $1)", comment.ID).Scan(&postUserID)
	if err != nil {
		http.Error(w, "Error fetching post information", http.StatusInternalServerError)
		return
	}

	if userID != postUserID {
		// If the user is not the author of the post, check if they're the author of the comment
		result, err := db.Exec("DELETE FROM comments WHERE id = $1 AND user_id = $2", comment.ID, userID)
		if err != nil {
			http.Error(w, "Error deleting comment", http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Comment not found or you don't have permission to delete it", http.StatusForbidden)
			return
		}
	} else {
		// If the user is the author of the post, they can delete any comment
		_, err := db.Exec("DELETE FROM comments WHERE id = $1", comment.ID)
		if err != nil {
			http.Error(w, "Error deleting comment", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func generateToken(userID int) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func getUserIDFromToken(r *http.Request) (int, error) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		return 0, fmt.Errorf("no token provided")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	return claims.UserID, nil
}

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	db := connectToDB()
	defer db.Close()

	err = db.QueryRow(
		"INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id",
		user.Username, user.Email, string(hashedPassword),
	).Scan(&user.ID)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	user.Password = "" // Don't send password back
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	db := connectToDB()
	defer db.Close()

	var user User
	err = db.QueryRow("SELECT id, password FROM users WHERE email = $1", credentials.Email).Scan(&user.ID, &user.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(user.ID)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"token": token, "user_id": user.ID})
}

func main() {
	http.HandleFunc("/", serveHTML)
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	http.HandleFunc("/users", getUsers)
	http.HandleFunc("/posts", getPosts)
	http.HandleFunc("/posts/create", createPost)
	http.HandleFunc("/comments/create", createComment)
	http.HandleFunc("/comments", getComments)
	http.HandleFunc("/posts/update", updatePost)
	http.HandleFunc("/posts/delete", deletePost)
	http.HandleFunc("/comments/update", updateComment)
	http.HandleFunc("/comments/delete", deleteComment)

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}
