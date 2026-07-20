package handlers

import (
	"context"
	"net/http"

	"github.com/Kurai1979/flashcard/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// sessionUserIDKey is the scs session key holding the logged-in user's id.
const sessionUserIDKey = "userID"

type contextKey string

const userContextKey contextKey = "user"

// RequireAuth blocks unauthenticated requests and loads the current user into
// the request context for downstream handlers. A stale or corrupt session
// (unparseable id, or a user that no longer exists) is destroyed and treated as
// logged out rather than a server error.
func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := h.Sessions.GetString(r.Context(), sessionUserIDKey)
		if userID == "" {
			h.redirectToLogin(w, r)
			return
		}

		var id pgtype.UUID
		if err := id.Scan(userID); err != nil {
			_ = h.Sessions.Destroy(r.Context())
			h.redirectToLogin(w, r)
			return
		}

		user, err := h.Queries.GetUserById(r.Context(), id)
		if err != nil {
			_ = h.Sessions.Destroy(r.Context())
			h.redirectToLogin(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// redirectToLogin sends the client to /login, using htmx's HX-Redirect header
// when the request came from htmx so the browser performs a full navigation.
func (h *Handler) redirectToLogin(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// userFromContext returns the user stashed by RequireAuth.
func userFromContext(ctx context.Context) (db.User, bool) {
	user, ok := ctx.Value(userContextKey).(db.User)
	return user, ok
}
