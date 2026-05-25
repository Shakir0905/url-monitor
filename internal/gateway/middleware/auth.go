package middleware

import (
	"context"
	"net/http"
	"strings"

	authpb "github.com/shakir/url-monitor/proto/auth"
)

type ctxKey string

const userIDKey ctxKey = "user_id"

// AuthMiddleware checks the Authorization header and injects user_id into the context.
func AuthMiddleware(authClient authpb.AuthServiceClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, `{"error":"missing or invalid token"}`, http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(header, "Bearer ")

			resp, err := authClient.ValidateToken(r.Context(), &authpb.ValidateTokenRequest{Token: token})
			if err != nil || !resp.GetValid() {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, resp.GetUserId())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the user_id set by AuthMiddleware.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(userIDKey).(int64)
	return id, ok
}
