package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Post struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type Comment struct {
	ID        int    `json:"id"`
	PostID    int    `json:"post_id"`
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
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

	rows, err := db.Query("SELECT id, username, email, password FROM users")
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Password)
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

func createPost(w http.ResponseWriter, r *http.Request) {
	db := connectToDB()
	defer db.Close()

	var post Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	post.UserID = 1

	err = db.QueryRow(
		"INSERT INTO posts (user_id, content) VALUES ($1, $2) RETURNING id",
		post.UserID, post.Content,
	).Scan(&post.ID)
	if err != nil {
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}

	row := db.QueryRow("SELECT username FROM users WHERE id = $1", post.UserID)
	err = row.Scan(&post.Username)
	if err != nil {
		http.Error(w, "Error fetching username", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// Получить все посты
func getPosts(w http.ResponseWriter, r *http.Request) {
	db := connectToDB()
	defer db.Close()

	rows, err := db.Query("SELECT posts.id, posts.user_id, posts.content, posts.created_at, users.username FROM posts JOIN users ON posts.user_id = users.id")
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

// Создать комментарий
func createComment(w http.ResponseWriter, r *http.Request) {
	db := connectToDB()
	defer db.Close()

	var comment Comment
	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	comment.UserID = 1

	err = db.QueryRow(
		"INSERT INTO comments (post_id, user_id, content) VALUES ($1, $2, $3) RETURNING id",
		comment.PostID, comment.UserID, comment.Content,
	).Scan(&comment.ID)
	if err != nil {
		http.Error(w, "Error creating comment", http.StatusInternalServerError)
		return
	}

	row := db.QueryRow("SELECT username FROM users WHERE id = $1", comment.UserID)
	err = row.Scan(&comment.Username)
	if err != nil {
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
// Редактировать пост
func updatePost(w http.ResponseWriter, r *http.Request) {
    db := connectToDB()
    defer db.Close()

    var post Post
    err := json.NewDecoder(r.Body).Decode(&post)
    if err != nil || post.ID == 0 {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    _, err = db.Exec("UPDATE posts SET content = $1 WHERE id = $2", post.Content, post.ID)
    if err != nil {
        http.Error(w, "Error updating post", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

// Удалить пост
func deletePost(w http.ResponseWriter, r *http.Request) {
    db := connectToDB()
    defer db.Close()

    var post Post
    err := json.NewDecoder(r.Body).Decode(&post)
    if err != nil || post.ID == 0 {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    _, err = db.Exec("DELETE FROM posts WHERE id = $1", post.ID)
    if err != nil {
        http.Error(w, "Error deleting post", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

// Редактировать комментарий
func updateComment(w http.ResponseWriter, r *http.Request) {
    db := connectToDB()
    defer db.Close()

    var comment Comment
    err := json.NewDecoder(r.Body).Decode(&comment)
    if err != nil || comment.ID == 0 {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    _, err = db.Exec("UPDATE comments SET content = $1 WHERE id = $2", comment.Content, comment.ID)
    if err != nil {
        http.Error(w, "Error updating comment", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

// Удалить комментарий
func deleteComment(w http.ResponseWriter, r *http.Request) {
    db := connectToDB()
    defer db.Close()

    var comment Comment
    err := json.NewDecoder(r.Body).Decode(&comment)
    if err != nil || comment.ID == 0 {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    _, err = db.Exec("DELETE FROM comments WHERE id = $1", comment.ID)
    if err != nil {
        http.Error(w, "Error deleting comment", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/", serveHTML)

	// Маршрут для статических файлов
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

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
