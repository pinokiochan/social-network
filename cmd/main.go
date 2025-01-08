package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"sync"

	"github.com/pinokiochan/social-network/internal/database"
	"github.com/pinokiochan/social-network/internal/handlers"
	"github.com/pinokiochan/social-network/internal/middleware"
	"github.com/pinokiochan/social-network/internal/logger"
	
)

var wg sync.WaitGroup

func main() {
	logger.Log.Info("Starting application")

	db, err := database.ConnectToDB()
	if err != nil {
		logger.Log.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	logger.Log.Info("Database connection established")

	userHandler := handlers.NewUserHandler(db)
	postHandler := handlers.NewPostHandler(db)
	commentHandler := handlers.NewCommentHandler(db)
	adminHandler := handlers.NewAdminHandler(db, &wg)

	mux := http.NewServeMux()

	// Configure routes
	fsWeb := http.FileServer(http.Dir("./web/static"))
	fsImg := http.FileServer(http.Dir("./web/img"))

	mux.Handle("/static/", http.StripPrefix("/static/", fsWeb))
	mux.Handle("/img/", http.StripPrefix("/img/", fsImg))

	// API routes configuration
	mux.HandleFunc("/api/register", userHandler.Register)
	mux.HandleFunc("/api/login", userHandler.Login)
	mux.HandleFunc("/api/users", middleware.JWT(userHandler.GetUsers))
	mux.HandleFunc("/api/posts", middleware.JWT(postHandler.GetPosts))
	mux.HandleFunc("/api/posts/create", middleware.JWT(postHandler.CreatePost))
	mux.HandleFunc("/api/posts/update", middleware.JWT(postHandler.UpdatePost))
	mux.HandleFunc("/api/posts/delete", middleware.JWT(postHandler.DeletePost))
	mux.HandleFunc("/api/comments", middleware.JWT(commentHandler.GetComments))
	mux.HandleFunc("/api/comments/create", middleware.JWT(commentHandler.CreateComment))
	mux.HandleFunc("/api/comments/update", middleware.JWT(commentHandler.UpdateComment))
	mux.HandleFunc("/api/comments/delete", middleware.JWT(commentHandler.DeleteComment))

	// Admin routes
	mux.HandleFunc("/admin", handlers.ServeAdminHTML)
	mux.HandleFunc("/api/admin/stats", middleware.AdminOnly(adminHandler.GetStats))
	mux.HandleFunc("/api/admin/broadcast-to-selected", adminHandler.BroadcastEmailToSelectedUsers)
	mux.HandleFunc("/api/admin/users", middleware.AdminOnly(adminHandler.GetUsers))
	mux.HandleFunc("/api/admin/users/delete", middleware.AdminOnly(adminHandler.DeleteUser))
	mux.HandleFunc("/api/admin/users/edit", middleware.AdminOnly(adminHandler.EditUser))

	mux.HandleFunc("/api/send-email", middleware.JWT(userHandler.SendEmail))
	mux.HandleFunc("/", handlers.ServeHTML)
	mux.HandleFunc("/email", handlers.ServeEmailHTML)

	srv := &http.Server{
		Addr:         "127.0.0.1:8080",
		Handler:      middleware.LoggingMiddleware(middleware.RateLimitMiddleware(mux)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		logger.Log.WithField("address", srv.Addr).Info("Starting server")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.WithError(err).Fatal("Server failed to start")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.WithError(err).Error("Server forced to shutdown")
	}

	logger.Log.Info("Waiting for background tasks to complete...")
	wg.Wait()

	logger.Log.Info("Server exited properly")
}

