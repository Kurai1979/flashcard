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

	// Re-read the list so the response can carry it back out-of-band; otherwise
	// the new card wouldn't show until a reload.
	cards, err := h.listFlashcards(r, user.ID)
	if err != nil {
		h.serverError(w, r, "list flashcards", err)
		return
	}

	// Hand back a cleared form so the user can add another card. Echoing the
	// saved card back instead would set an ID on the form and turn the create
	// widget into an editor for that one card.
	h.render(w, r, http.StatusOK, templates.FlashcardCreated(cards))
}

// RemoveFlashcard deletes one of the caller's cards. The delete button targets
// the card's own <li> with hx-swap="outerHTML", so an empty 200 body drops the
// row out of the list.
func (h *Handler) RemoveFlashcard(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireUser(w, r)
	if !ok {
		return
	}

	var id pgtype.UUID
	if err := id.Scan(chi.URLParam(r, "id")); err != nil {
		http.NotFound(w, r)
		return
	}

	c, err := h.Queries.DeleteFlashcard(r.Context(), db.DeleteFlashcardParams{
		ID:     id,
		UserID: user.ID,
	})
	if err != nil {
		h.serverError(w, r, "delete flashcard", err)
		return
	}

	// The query is scoped to the caller, so a zero count means the card either
	// doesn't exist or isn't theirs. 404 covers both without saying which.
	if c == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// flashcardPageSize is how many cards one page of the list holds. ListFlashcards
// pages by cursor, so this is the first page; the rest are reachable by passing
// the last row's ID back as the cursor.
const flashcardPageSize = 50

// listFlashcards reads the first page of a user's cards. Both the decks page and
// the create-card response render the same list, so the paging arguments live in
// one place rather than being repeated at each call site.
func (h *Handler) listFlashcards(r *http.Request, userID pgtype.UUID) ([]db.ListFlashcardsRow, error) {
	return h.Queries.ListFlashcards(r.Context(), db.ListFlashcardsParams{
		UserID: userID,
		// A zero-value UUID is invalid, which the query reads as SQL NULL: no
		// cursor, so start at the newest card.
		Cursor:   pgtype.UUID{},
		PageSize: flashcardPageSize,
	})
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

func (h *Handler) DeleteFlashcard(w http.ResponseWriter, r *http.Request) {
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

	c, err := h.Queries.DeleteFlashcard(r.Context(), db.DeleteFlashcardParams{
		ID:     id,
		UserID: user.ID,
	})

	if err != nil {
		h.serverError(w, r, "delete flashcard", err)
		return
	}

	if c == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
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
