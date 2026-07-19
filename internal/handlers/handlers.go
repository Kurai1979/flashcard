package handlers

import (
	"log/slog"
	"net/http"

	"github.com/Kurai1979/flashcard/internal/db"
	"github.com/a-h/templ"
)

type Handler struct {
	Queries db.Querier
	Logger  *slog.Logger
}

func New(queries db.Querier, logger *slog.Logger) *Handler {
	return &Handler{
		Queries: queries,
		Logger:  logger,
	}
}

func (h *Handler) serverError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	h.Logger.ErrorContext(r.Context(), msg, "err", err)
	http.Error(w, "server error", http.StatusInternalServerError)
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request, status int, c templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := c.Render(r.Context(), w); err != nil {
		h.Logger.ErrorContext(r.Context(), "render template", "err", err)
	}
}
