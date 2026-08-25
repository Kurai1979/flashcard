package handlers

import (
	"net/http"

	"github.com/Kurai1979/flashcard/internal/db"
	"github.com/Kurai1979/flashcard/internal/templates"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *Handler) AddDeck(w http.ResponseWriter, r *http.Request) {

	user, ok := h.requireUser(w, r)
	if !ok {
		return
	}

	if !h.parseForm(w, r) {
		return
	}

	form := parseDeckform(r)

	// htmx only swaps content on 2xx responses, so validation errors return 200
	// with the re-rendered form rather than a 4xx the client would ignore.
	if msg := validateDeckform(form); msg != "" {
		form.Error = msg
		h.render(w, r, http.StatusOK, templates.SaveDeck(form))
		return
	}

	_, err := h.Queries.CreateDeck(r.Context(), db.CreateDeckParams{

		UserID:      user.ID,
		Name:        form.Name,
		Description: textOrNull(form.Description),
	})

	if err != nil {
		h.serverError(w, r, "create deck", err)
		return
	}

	h.render(w, r, http.StatusCreated, templates.SaveDeck(templates.Deckform{Notice: "deck created"}))
}

// GetDecks renders the deck/card authoring page together with the caller's own
// decks. The card form it contains posts to /flashcard on its own.
func (h *Handler) GetDecks(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireUser(w, r)
	if !ok {
		return
	}

	decks, err := h.Queries.ListDecks(r.Context(), user.ID)
	if err != nil {
		h.serverError(w, r, "list decks", err)
		return
	}

	h.render(w, r, http.StatusOK, templates.DecksPage(user.Email, decks))
}

// RemoveDeck deletes one of the caller's decks. The delete button targets the
// deck's own <li> with hx-swap="outerHTML", so an empty 200 body drops the row
// out of the list.
func (h *Handler) RemoveDeck(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireUser(w, r)
	if !ok {
		return
	}

	var id pgtype.UUID
	if err := id.Scan(chi.URLParam(r, "id")); err != nil {
		http.NotFound(w, r)
		return
	}

	// TODO: delete the deck for id + user.ID and 404 when nothing was removed.

	c, err := h.Queries.DeleteDeck(r.Context(), db.DeleteDeckParams{
		ID:     id,
		UserID: user.ID,
	})

	if err != nil {
		h.serverError(w, r, "delete deck", err)
		return
	}

	if c == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func validateDeckform(deckform templates.Deckform) string {
	if deckform.Name != "" {
		return ""
	}
	return "Name required"
}

func parseDeckform(r *http.Request) templates.Deckform {
	return templates.Deckform{
		ID:          chi.URLParam(r, "id"),
		Name:        formString(r, "name"),
		Description: formString(r, "description"),
	}
}
