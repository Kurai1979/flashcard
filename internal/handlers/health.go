package handlers

import (
	"net/http"
)

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if _, err := h.Queries.GetUserByEmail(r.Context(), "healthcheck@example.invalid"); err != nil {
		h.Logger.DebugContext(r.Context(), "health check db ping returned error (expected for missing user)", "err", err)
		h.serverError(w, r, "server down", err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
