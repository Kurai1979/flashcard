package handlers

import (
	"net/http"

	"github.com/Kurai1979/flashcard/internal/templates"
	"github.com/justinas/nosurf"
)

// CSRF returns middleware that wraps handlers with nosurf CSRF protection. Safe
// methods (GET/HEAD/OPTIONS) pass through; state-changing requests must present
// a token (form field csrf_token or header X-CSRF-Token) matching the per-browser
// base cookie, otherwise they get a 400. secure marks the base cookie Secure and
// should be true whenever the site is served over HTTPS.
func (h *Handler) CSRF(secure bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		csrfHandler := nosurf.New(next)
		csrfHandler.SetBaseCookie(http.Cookie{
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   secure,
		})
		// nosurf's Origin check derives the expected scheme from isTLS, which
		// defaults to always-true. Tie it to the real connection so the Origin
		// header matches over plain HTTP locally and HTTPS in production (incl.
		// behind a TLS-terminating proxy that sets X-Forwarded-Proto).
		csrfHandler.SetIsTLSFunc(func(r *http.Request) bool {
			return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
		})
		csrfHandler.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.Logger.WarnContext(r.Context(), "csrf validation failed",
				"path", r.URL.Path, "reason", nosurf.Reason(r),
				"origin", r.Header.Get("Origin"), "host", r.Host)
			http.Error(w, "invalid CSRF token", http.StatusBadRequest)
		}))
		return csrfHandler
	}
}

// CSRFToken publishes the current request's masked CSRF token into the context
// so form templates can embed it. Must run after CSRF.
func (h *Handler) CSRFToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := templates.WithCSRFToken(r.Context(), nosurf.Token(r))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
