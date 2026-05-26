package handlers

import (
	"log/slog"
	"net/http"

	"github.com/Kurai1979/flashcard/internal/db"
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
