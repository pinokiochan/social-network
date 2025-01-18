package middleware

import (
	"github.com/pinokiochan/social-network/internal/auth"
	"net/http"
	"strings"
)

// AdminOnly проверяет, является ли пользователь администратором
func AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем токен из заголовка Authorization
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
			return
		}

		// Проверяем, что токен начинается с "Bearer "
		if !strings.HasPrefix(token, "Bearer ") {
			http.Error(w, "Unauthorized: Invalid token format", http.StatusUnauthorized)
			return
		}

		// Обрезаем "Bearer " и проверяем токен
		token = token[len("Bearer "):]

		// Проверяем и декодируем токен
		claims, err := auth.VerifyToken(token)
		if err != nil {
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		// Проверяем, является ли пользователь администратором
		if !claims.IsAdmin {
			http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
			return
		}

		// Если все проверки пройдены, продолжаем выполнение запроса
		next.ServeHTTP(w, r)
	}
}
