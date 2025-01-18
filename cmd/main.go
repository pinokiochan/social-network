package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pinokiochan/social-network/internal/database"
	"github.com/pinokiochan/social-network/internal/handlers"
	"github.com/pinokiochan/social-network/internal/logger"
	"github.com/pinokiochan/social-network/internal/middleware"
	"github.com/sirupsen/logrus"
)

var wg sync.WaitGroup

func main() {
	// Открытие/создание файла для логирования
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logger.Log.Fatal("Failed to open log file:", err)
	}
	defer logFile.Close()

	// Настройка логгера для записи в файл
	logger.Log.SetOutput(logFile)                    // Все логи теперь будут записываться в файл
	logger.Log.SetFormatter(&logrus.JSONFormatter{}) // Форматируем логи в JSON

	// Логируем начало работы приложения
	logger.Log.Info("Starting application")

	// Подключение к базе данных
	db, err := database.ConnectToDB()
	if err != nil {
		logger.Log.WithError(err).Fatal("Failed to connect to database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Log.WithError(err).Error("Failed to close database connection")
		}
	}()

	logger.Log.Info("Database connection established")

	// Инициализация обработчиков
	userHandler := handlers.NewUserHandler(db)
	postHandler := handlers.NewPostHandler(db)
	commentHandler := handlers.NewCommentHandler(db)
	adminHandler := handlers.NewAdminHandler(db, &wg)

	// Создание нового ServeMux (роутера)
	mux := http.NewServeMux()

	// Обслуживание статичных файлов
	fsWeb := http.FileServer(http.Dir("./web/static"))
	fsImg := http.FileServer(http.Dir("./web/img"))
	mux.Handle("/static/", http.StripPrefix("/static/", fsWeb))
	mux.Handle("/img/", http.StripPrefix("/img/", fsImg))

	// Настройка API-роутов
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

	// Админ-роуты (ограничены пользователями с правами администратора)
	mux.HandleFunc("/admin", handlers.ServeAdminHTML)
	mux.HandleFunc("/api/admin/stats", middleware.AdminOnly(adminHandler.GetStats))
	mux.HandleFunc("/api/admin/broadcast-to-selected", adminHandler.BroadcastEmailToSelectedUsers)
	mux.HandleFunc("/api/admin/users", middleware.AdminOnly(adminHandler.GetUsers))
	mux.HandleFunc("/api/admin/users/delete", middleware.AdminOnly(adminHandler.DeleteUser))
	mux.HandleFunc("/api/admin/users/edit", middleware.AdminOnly(adminHandler.EditUser))

	// Регулярные HTML-страницы
	mux.HandleFunc("/", handlers.ServeHTML)
	mux.HandleFunc("/email", handlers.ServeEmailHTML)

	// Создание и настройка HTTP-сервера с тайм-аутами
	port := os.Getenv("APP_PORT") // Render предоставляет эту переменную окружения
	if port == "" {
		port = "8080" // default to 8080 if no PORT variable is set
	}

	srv := &http.Server{
		Addr:         ":" + port, // Используем динамический порт
		Handler:      middleware.LoggingMiddleware(middleware.RateLimitMiddleware(mux)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Запуск сервера в горутине, чтобы он не блокировал основное выполнение
	go func() {
		logger.Log.WithField("address", srv.Addr).Info("Starting server")

		// Запуск сервера и логирование ошибок
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.WithError(err).Fatal("Server failed to start")
		}
	}()

	// Ожидание сигнала завершения (Ctrl+C или системное завершение)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Блокируем выполнение до получения сигнала

	// Логируем процесс завершения работы
	logger.Log.Info("Server is shutting down...")

	// Создание контекста с тайм-аутом, чтобы завершить текущие запросы
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Грамотное завершение работы сервера
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.WithError(err).Error("Server forced to shutdown")
	}

	// Ожидание завершения фоновых задач перед полным завершением
	logger.Log.Info("Waiting for background tasks to complete...")
	wg.Wait()

	// Логируем успешное завершение работы
	logger.Log.Info("Server exited properly")
}
