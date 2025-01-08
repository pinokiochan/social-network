package middleware

import (
	"github.com/pinokiochan/social-network/internal/auth"
	"net/http"
)

// AdminOnly проверяет, является ли пользователь администратором
func AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Проверяем и декодируем токен
		claims, err := auth.VerifyToken(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Проверяем, является ли пользователь администратором
		if !claims.IsAdmin {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}
		

		// Если все проверки пройдены, продолжаем выполнение запроса
		next.ServeHTTP(w, r)

	}
	
}
