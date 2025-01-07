package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	adminHandler := handlers.NewAdminHandler(db)

	// Create new ServeMux for better control
	mux := http.NewServeMux()

	// Static file servers
	fsWeb := http.FileServer(http.Dir("./web/static"))
	fsImg := http.FileServer(http.Dir("./web/img"))

	// Routes
	mux.Handle("/static/", http.StripPrefix("/static/", fsWeb))
	mux.Handle("/img/", http.StripPrefix("/img/", fsImg))

	// API routes
	mux.HandleFunc("/api/register", userHandler.Register)
	mux.HandleFunc("/api/login", userHandler.Login)

	// Protected routes with JWT middleware
	mux.HandleFunc("/api/users", middleware.JWT(userHandler.GetUsers))
	mux.HandleFunc("/api/posts", middleware.JWT(postHandler.GetPosts))
	mux.HandleFunc("/api/posts/create", middleware.JWT(postHandler.CreatePost))
	mux.HandleFunc("/api/posts/update", middleware.JWT(postHandler.UpdatePost))
	mux.HandleFunc("/api/posts/delete", middleware.JWT(postHandler.DeletePost))
	mux.HandleFunc("/api/comments", middleware.JWT(commentHandler.GetComments))
	mux.HandleFunc("/api/comments/create", middleware.JWT(commentHandler.CreateComment))
	mux.HandleFunc("/api/comments/update", middleware.JWT(commentHandler.UpdateComment))
	mux.HandleFunc("/api/comments/delete", middleware.JWT(commentHandler.DeleteComment))

	// Admin-specific routes with AdminOnly middleware
	mux.HandleFunc("/admin", handlers.ServeAdminHTML)                               // This will serve HTML content for the admin panel
	mux.HandleFunc("/api/admin/stats", middleware.AdminOnly(adminHandler.GetStats)) // Admin protected
	// Inside main function
	mux.HandleFunc("/api/admin/broadcast-to-selected", adminHandler.BroadcastEmailToSelectedUsers) // Admin protected

	mux.HandleFunc("/api/admin/users", middleware.AdminOnly(adminHandler.GetUsers))          // Admin protected
	mux.HandleFunc("/api/admin/users/delete", middleware.AdminOnly(adminHandler.DeleteUser)) // Admin protected
	mux.HandleFunc("/api/admin/users/edit", middleware.AdminOnly(adminHandler.EditUser))     // Admin protected

	// New email route
	mux.HandleFunc("/api/send-email", middleware.JWT(userHandler.SendEmail))

	// Serve frontend
	mux.HandleFunc("/", handlers.ServeHTML)           // Главная страница
	mux.HandleFunc("/email", handlers.ServeEmailHTML) // Email Broadcast

	// Create server with timeouts
	srv := &http.Server{
		Addr:         "127.0.0.1:8080",
		Handler:      middleware.LoggingMiddleware(middleware.RateLimitMiddleware(mux)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		fmt.Println("Server is running on http://127.0.0.1:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
