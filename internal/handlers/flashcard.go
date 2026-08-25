package handlers

import (
	"net/http"

	"github.com/Kurai1979/flashcard/internal/db"
	"github.com/Kurai1979/flashcard/internal/templates"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) CreateFlashcard(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireUser(w, r)
	if !ok {
		return
	}

	if !h.parseForm(w, r) {
		return
	}

	form := parseFlashcardForm(r)

	// htmx only swaps content on 2xx responses, so validation errors return 200
	// with the re-rendered form rather than a 4xx the client would ignore.
	if msg := validateFlashcard(form); msg != "" {
		form.Error = msg
		h.render(w, r, http.StatusOK, templates.SaveFlashcard(form))
		return
	}

	_, err := h.Queries.CreateFlashcard(r.Context(), db.CreateFlashcardParams{
		UserID:  user.ID,
		Front:   form.Front,
		Back:    form.Back,
		Example: textOrNull(form.Example),
	})
	if err != nil {
		h.serverError(w, r, "create flashcard", err)
		return
	}

	// Hand back a cleared form so the user can add another card. Echoing the
	// saved card back instead would set an ID on the form and turn the create
	// widget into an editor for that one card.
	h.render(w, r, http.StatusOK, templates.SaveFlashcard(templates.FlashcardForm{
		Notice: "Card added.",
	}))
}

func (h *Handler) UpdateFlashcard(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireUser(w, r)
	if !ok {
		return
	}

	if !h.parseForm(w, r) {
		return
	}

	var id pgtype.UUID
	if err := id.Scan(chi.URLParam(r, "id")); err != nil {
		http.NotFound(w, r)
		return
	}

	form := parseFlashcardForm(r)

	if msg := validateFlashcard(form); msg != "" {
		form.Error = msg
		h.render(w, r, http.StatusOK, templates.SaveFlashcard(form))
		return
	}

	card, err := h.Queries.UpdateFlashcard(r.Context(), db.UpdateFlashcardParams{
		Front:   form.Front,
		Back:    form.Back,
		Example: textOrNull(form.Example),
		ID:      id,
		UserID:  user.ID,
	})

	if !h.queryOK(w, r, "update flashcard", err) {
		return
	}

	h.render(w, r, http.StatusOK, templates.SaveFlashcard(templates.FlashcardForm{
		ID:      card.ID.String(),
		Front:   card.Front,
		Back:    card.Back,
		Example: textOrEmpty(card.Example),
	}))
}

func parseFlashcardForm(r *http.Request) templates.FlashcardForm {
	return templates.FlashcardForm{
		ID:      chi.URLParam(r, "id"),
		Front:   formString(r, "front"),
		Back:    formString(r, "back"),
		Example: formString(r, "example"),
	}
}

// validateFlashcard returns an empty string when the form is usable.
func validateFlashcard(f templates.FlashcardForm) string {
	switch {
	case f.Front == "":
		return "front is required"
	case f.Back == "":
		return "back is required"
	}
	return ""
}
