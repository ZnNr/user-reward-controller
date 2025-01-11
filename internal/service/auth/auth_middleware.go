package auth

import (
	"github.com/ZnNr/user-reward-controller/internal/errors"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"strings"
)

// AuthMiddleware проверяет JWT токен
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Извлечение токена из заголовков
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, errors.ErrMsgInvalidToken, http.StatusUnauthorized)
			return
		}

		tokenString = strings.TrimSpace(strings.TrimPrefix(tokenString, "Bearer "))

		// Парсинг и валидация токена
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Проверка алгоритма токена
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.NewInvalidToken("token uses invalid signing method", nil)
			}
			return []byte("your_secret_key"), nil
		})

		if err != nil {
			// Проверка на недействительный токен
			if errors.IsInvalidToken(err) {
				http.Error(w, errors.ErrMsgInvalidToken, http.StatusUnauthorized)
				return
			}
			http.Error(w, "Unauthorized: token parsing error", http.StatusUnauthorized)
			return
		}

		// Проверка токена на валидность
		if token.Valid {
			// Токен валиден, продолжаем выполнение следующего обработчика
			next.ServeHTTP(w, r)
			return
		} else {
			http.Error(w, errors.ErrMsgInvalidToken, http.StatusUnauthorized)
			return
		}
	})
}
