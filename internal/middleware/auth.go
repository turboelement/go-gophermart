package middleware

import (
	"net/http"

	"go-gophermart/internal/auth"

	"go.uber.org/zap"
)

func AuthMiddleware(secretKey string, logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := auth.GetUserIDFromCookie(r, secretKey)

			if err != nil {
				logger.Debug("Authentication error", zap.Error(err))
				ctx := auth.SetUserIDInContext(r.Context(), "")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			ctx := auth.SetUserIDInContext(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(r *http.Request) (string, error) {
	return auth.GetUserIDFromContext(r.Context())
}
