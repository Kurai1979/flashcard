package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Kurai1979/flashcard/internal/db"
	"github.com/a-h/templ"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const maxFormBytes = 1 << 20

type Handler struct {
	Queries  db.Querier
	Logger   *slog.Logger
	Sessions *scs.SessionManager
}

func New(queries db.Querier, logger *slog.Logger, sessions *scs.SessionManager) *Handler {
	return &Handler{
		Queries:  queries,
		Logger:   logger,
		Sessions: sessions,
	}
}

func (h *Handler) serverError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	h.Logger.ErrorContext(r.Context(), msg, "err", err)
	http.Error(w, "server error", http.StatusInternalServerError)
}

// queryOK reports whether a query succeeded, writing the response itself when
// it didn't. Queries scoped to the current user return ErrNoRows both for rows
// that don't exist and for rows owned by someone else; 404 covers both without
// confirming which, and neither is a server fault worth logging as one.
func (h *Handler) queryOK(w http.ResponseWriter, r *http.Request, msg string, err error) bool {
	switch {
	case err == nil:
		return true
	case errors.Is(err, pgx.ErrNoRows):
		http.NotFound(w, r)
	default:
		h.serverError(w, r, msg, err)
	}
	return false
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request, status int, c templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := c.Render(r.Context(), w); err != nil {
		h.Logger.ErrorContext(r.Context(), "render template", "err", err)
	}
}

func (h *Handler) parseForm(w http.ResponseWriter, r *http.Request) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBytes)
	if err := r.ParseForm(); err != nil {
		h.Logger.WarnContext(r.Context(), "parse form", "err", err)
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return false
	}
	return true
}

func (h *Handler) requireUser(w http.ResponseWriter, r *http.Request) (db.User, bool) {
	user, ok := userFromContext(r.Context())
	if !ok {
		h.redirectToLogin(w, r)
	}
	return user, ok
}

// formString returns the trimmed value of a submitted field.
func formString(r *http.Request, key string) string {
	return strings.TrimSpace(r.PostFormValue(key))
}

// textOrNull maps empty input to SQL NULL so "no value" has a single
// representation in the database rather than both NULL and an empty string.
func textOrNull(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

// textOrEmpty is the inverse of textOrNull: SQL NULL renders as an empty field.
func textOrEmpty(text pgtype.Text) string {
	if !text.Valid {
		return ""
	}
	return text.String
}
