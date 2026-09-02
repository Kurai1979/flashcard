package handlers

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"uuid"

	"github.com/Kurai1979/flashcard/internal/auth"
	"github.com/Kurai1979/flashcard/internal/templates"
	"github.com/jackc/pgx/v5"
)

// LoginPage renders the login form, or bounces already-authenticated users
// straight to the dashboard.
func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if h.Sessions.GetString(r.Context(), sessionUserIDKey) != "" {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	h.render(w, r, http.StatusOK, templates.LoginPage())
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if !h.parseForm(w, r) {
		return
	}

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	// Every failure returns the same message (200 so htmx swaps the fragment) to
	// avoid revealing whether an email is registered.
	const genericErr = "invalid email or password"
	formError := func() {
		h.render(w, r, http.StatusOK, templates.Login(templates.LoginForm{
			Email: email,
			Error: genericErr,
		}))
	}

	e, err := mail.ParseAddress(email)
	if err != nil {
		formError()
		return
	}
	normalizedEmail := strings.ToLower(strings.TrimSpace(e.Address))

	user, err := h.Queries.GetUserByEmail(r.Context(), normalizedEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			formError()
			return
		}
		h.serverError(w, r, "lookup user", err)
		return
	}

	// A malformed stored hash or an out-of-range password surfaces as an error;
	// treat it as an authentication failure rather than a 500.
	ok, err := auth.AuthorizeUser(user.PasswordHash, password)
	if err != nil || !ok || !user.IsActive {
		formError()
		return
	}

	// Rotate the session token now that the privilege level changes, to defeat
	// session fixation.
	if err := h.Sessions.RenewToken(r.Context()); err != nil {
		h.serverError(w, r, "renew session", err)
		return
	}
	h.Sessions.Put(r.Context(), sessionUserIDKey, uuid.UUID(user.ID.Bytes).String())

	if err := h.Queries.UpdateLastLogin(r.Context(), user.ID); err != nil {
		// Non-fatal: the login already succeeded.
		h.Logger.WarnContext(r.Context(), "update last login", "err", err)
	}

	h.Logger.InfoContext(r.Context(), "user logged in", "email", user.Email)
	h.redirectAfterAuth(w, r, "/dashboard")
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.Sessions.Destroy(r.Context()); err != nil {
		h.serverError(w, r, "destroy session", err)
		return
	}
	h.redirectAfterAuth(w, r, "/login")
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		h.redirectToLogin(w, r)
		return
	}
	h.render(w, r, http.StatusOK, templates.DashboardPage(user.Email))
}

// redirectAfterAuth navigates to to, honoring htmx's HX-Redirect header when the
// request originated from htmx.
func (h *Handler) redirectAfterAuth(w http.ResponseWriter, r *http.Request, to string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", to)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, to, http.StatusSeeOther)
}
