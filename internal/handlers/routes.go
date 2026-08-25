package handlers

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Routes builds the application router: global middleware, unauthenticated
// endpoints, and the session/CSRF-protected group with its authenticated
// sub-group. secureCookies marks the CSRF cookie Secure for HTTPS.
func (h *Handler) Routes(sessionManager *scs.SessionManager, secureCookies bool) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Static assets and health checks don't need a session.
	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))
	r.Get("/healthz", h.Health)

	// Root lands on the dashboard, which itself redirects to /login when the
	// visitor isn't authenticated.
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	})

	// Everything else loads/saves the session cookie and is CSRF-protected.
	r.Group(func(r chi.Router) {
		r.Use(sessionManager.LoadAndSave)
		r.Use(h.CSRF(secureCookies))
		r.Use(h.CSRFToken)

		r.Get("/login", h.LoginPage)
		r.Post("/login", h.Login)
		r.Post("/logout", h.Logout)
		r.Get("/users/new", h.NewUser)
		r.Post("/users", h.CreateUser)

		// Authenticated area.
		r.Group(func(r chi.Router) {
			r.Use(h.RequireAuth)
			r.Get("/dashboard", h.Dashboard)
			r.Get("/decks", h.GetDecks)
			r.Post("/decks", h.AddDeck)
			r.Delete("/decks/{id}", h.RemoveDeck)
			r.Post("/flashcard", h.CreateFlashcard)
			r.Post("/flashcard/{id}", h.UpdateFlashcard)
		})
	})

	return r
}
