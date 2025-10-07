package apiserver

import (
	"context"
	"go-sqs/store"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func NewLoggerMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("http request", "method", r.Method, "path", r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}

type userCtxKey struct{}

func ContextWithUser(r *http.Request, user *store.User) context.Context {
	return context.WithValue(r.Context(), userCtxKey{}, user)
}

func NewAuthMiddleware(JwtManager *JwtManager, userStore *store.UserStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/auth/") {
				next.ServeHTTP(w, r)
				return
			}

			authorizationHeader := r.Header.Get("Authorization")
			var token string
			if parts := strings.Split(authorizationHeader, "Bearer "); len(parts) == 2 {
				token = parts[1]
			}

			if token == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			parsedToken, err := JwtManager.Parse(token)
			if err != nil {
				slog.Error("faild to parse token", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if !JwtManager.IsAccessToken(parsedToken) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("not an access token"))
				return
			}

			userIdStr, err := parsedToken.Claims.GetSubject()
			if err != nil {
				slog.Error("faild to get subject from token", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			userId, err := uuid.Parse(userIdStr)
			if err != nil {
				slog.Error("faild to parse user id", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			user, err := userStore.ByID(r.Context(), userId)
			if err != nil {
				slog.Error("faild to get user by id", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(ContextWithUser(r, user)))
		})
	}
}
