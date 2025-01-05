package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pinokiochan/social-network/internal/database"
	"github.com/pinokiochan/social-network/internal/handlers"
	"github.com/pinokiochan/social-network/internal/middleware"
)

func main() {
	// Initialize database connection
	db := database.ConnectToDB()
	defer db.Close()

	// Initialize handlers
	userHandler := handlers.NewUserHandler(db)
	postHandler := handlers.NewPostHandler(db)
	commentHandler := handlers.NewCommentHandler(db)

	// Static file servers
	fsWeb := http.FileServer(http.Dir("./web/static"))
	fsImg := http.FileServer(http.Dir("./web/img"))

	// Routes
	http.Handle("/static/", http.StripPrefix("/static/", fsWeb))
	http.Handle("/img/", http.StripPrefix("/img/", fsImg))

	// API routes
	http.HandleFunc("/api/register", userHandler.Register)
	http.HandleFunc("/api/login", userHandler.Login)

	// Protected routes
	http.HandleFunc("/api/users", middleware.JWT(userHandler.GetUsers))
	http.HandleFunc("/api/posts", middleware.JWT(postHandler.GetPosts))
	http.HandleFunc("/api/posts/create", middleware.JWT(postHandler.CreatePost))
	http.HandleFunc("/api/posts/update", middleware.JWT(postHandler.UpdatePost))
	http.HandleFunc("/api/posts/delete", middleware.JWT(postHandler.DeletePost))
	http.HandleFunc("/api/comments", middleware.JWT(commentHandler.GetComments))
	http.HandleFunc("/api/comments/create", middleware.JWT(commentHandler.CreateComment))
	http.HandleFunc("/api/comments/update", middleware.JWT(commentHandler.UpdateComment))
	http.HandleFunc("/api/comments/delete", middleware.JWT(commentHandler.DeleteComment))

	// Serve frontend
	http.HandleFunc("/", handlers.ServeHTML)

	fmt.Println("Server is running on http://127.0.0.1:8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}
