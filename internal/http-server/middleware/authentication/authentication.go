package authentication

import (
	"context"
	"log/slog"
	"net/http"

	resp "url-shorter/internal/lib/api/response"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	resp.Response
}

type UserAuth interface {
	ValidateUser(username, password string) (bool, error)
}

type Request struct {
    Username   string `json:"username" validate:"required,min=3,max=50,alphanum"`
    Password   string `json:"password" validate:"required,min=8"`
}

type contextKey string

const usernameKey contextKey = "username"


func BasicAuthMiddleware(log *slog.Logger, userAuth UserAuth) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const fn = "middleware.authentication.BasicAuthMiddleware"
			
			log = log.With(
				slog.String("fn", fn),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)

			username, password, ok := r.BasicAuth()
			if !ok {
				log.Warn("missing or invalid Authorization header", slog.String("fn", fn))
				w.Header().Set("WWW-Authenticate", `Basic realm="url-shorter"`)
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w,r, resp.Error("Unauthorized"))
				return
			}
			
			ok, err := userAuth.ValidateUser(username, password)
			if err != nil {
				log.Error("%s: %w", fn, err)
				log.Warn("missing or invalid Authorization header", slog.String("fn", fn))
				w.Header().Set("WWW-Authenticate", `Basic realm="url-shorter"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !ok {
				log.Error("%s: %w", fn, err)
				log.Warn("invalid credentials", slog.String("username", username))
				w.Header().Set("WWW-Authenticate", `Basic realm="url-shorter"`)
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w,r, resp.Error("Unauthorized"))
				return
			}

			ctx := context.WithValue(r.Context(), usernameKey, username)
			r = r.WithContext(ctx)

			h.ServeHTTP(w, r)
		})
	}
}
