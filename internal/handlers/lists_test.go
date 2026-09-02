package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"uuid"

	"github.com/Kurai1979/flashcard/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// serveAs runs a request through handler with user already in the context, the
// way RequireAuth would have put it there.
func serveAs(handler http.HandlerFunc, user db.User, method, target string, form url.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	req = req.WithContext(context.WithValue(req.Context(), userContextKey, user))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func newTestUUID() pgtype.UUID {
	return pgtype.UUID{Bytes: uuid.New(), Valid: true}
}

// serveDelete runs RemoveFlashcard with the {id} route param populated the way
// chi would have, plus the signed-in user in the context.
func serveDelete(h *Handler, user db.User, target, id string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, target, nil)
	req.Header.Set("HX-Request", "true")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	ctx = context.WithValue(ctx, userContextKey, user)

	rec := httptest.NewRecorder()
	h.RemoveFlashcard(rec, req.WithContext(ctx))
	return rec
}

// assertOOBTag checks that the element carrying id is the same one marked for
// the out-of-band swap. Asserting both strings appear somewhere in the body
// would pass even if htmx had nothing to act on, which is the failure mode this
// is guarding against.
func assertOOBTag(t *testing.T, body, id string) {
	t.Helper()
	start := strings.Index(body, `id="`+id+`"`)
	if start == -1 {
		t.Fatalf("no element with id %q in the response", id)
	}
	end := strings.Index(body[start:], ">")
	if end == -1 {
		t.Fatalf("unterminated tag for id %q", id)
	}
	if !strings.Contains(body[start:start+end], `hx-swap-oob="true"`) {
		t.Errorf("element %q is not marked hx-swap-oob; htmx will ignore it", id)
	}
}

// A created card has to come back in the same response as the cleared form,
// marked for an out-of-band swap. Without that fragment the card only appears
// after a manual reload, which is the bug this guards against.
func TestCreateFlashcard_ReturnsListOutOfBand(t *testing.T) {
	user := newTestUser(t, "user@example.com", "correct-horse-battery", true)
	var listed bool
	q := &stubQuerier{
		createFlashcard: func(_ context.Context, arg db.CreateFlashcardParams) (db.Flashcard, error) {
			if arg.Front != "einen Vertrag abschließen" {
				t.Fatalf("unexpected front: %q", arg.Front)
			}
			if arg.UserID != user.ID {
				t.Fatal("card was not scoped to the signed-in user")
			}
			return db.Flashcard{ID: newTestUUID(), Front: arg.Front, Back: arg.Back}, nil
		},
		listFlashcards: func(_ context.Context, arg db.ListFlashcardsParams) ([]db.ListFlashcardsRow, error) {
			listed = true
			if arg.UserID != user.ID {
				t.Fatal("list was not scoped to the signed-in user")
			}
			if arg.Cursor.Valid {
				t.Error("first page should be requested with a NULL cursor")
			}
			if arg.PageSize != flashcardPageSize {
				t.Errorf("page size = %d, want %d", arg.PageSize, flashcardPageSize)
			}
			return []db.ListFlashcardsRow{
				{ID: newTestUUID(), Front: "einen Vertrag abschließen", Back: "to conclude a contract"},
			}, nil
		},
	}
	h, _ := newTestHandler(q)

	rec := serveAs(h.CreateFlashcard, user, http.MethodPost, "/flashcard", url.Values{
		"front": {"einen Vertrag abschließen"},
		"back":  {"to conclude a contract"},
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !listed {
		t.Fatal("handler did not re-read the card list")
	}

	body := rec.Body.String()
	for _, want := range []string{
		`Card added.`,            // the cleared form came back
		"to conclude a contract", // and the list, containing the new card
	} {
		if !strings.Contains(body, want) {
			t.Errorf("response is missing %q", want)
		}
	}
	assertOOBTag(t, body, "flashcard-list")
	// The umlaut has to survive templ's escaping intact.
	if !strings.Contains(body, "einen Vertrag abschlie") {
		t.Error("card front missing from the rendered list")
	}
}

// The same contract for decks: creating one refreshes the list in place.
func TestAddDeck_ReturnsListOutOfBand(t *testing.T) {
	user := newTestUser(t, "user@example.com", "correct-horse-battery", true)
	q := &stubQuerier{
		createDeck: func(_ context.Context, arg db.CreateDeckParams) (db.Deck, error) {
			if arg.UserID != user.ID {
				t.Fatal("deck was not scoped to the signed-in user")
			}
			return db.Deck{ID: newTestUUID(), Name: arg.Name}, nil
		},
		listDecks: func(_ context.Context, id pgtype.UUID) ([]db.Deck, error) {
			if id != user.ID {
				t.Fatal("list was not scoped to the signed-in user")
			}
			return []db.Deck{{ID: newTestUUID(), Name: "German B2"}}, nil
		},
	}
	h, _ := newTestHandler(q)

	rec := serveAs(h.AddDeck, user, http.MethodPost, "/decks", url.Values{
		"name": {"German B2"},
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	body := rec.Body.String()
	for _, want := range []string{`deck created`, "German B2"} {
		if !strings.Contains(body, want) {
			t.Errorf("response is missing %q", want)
		}
	}
	assertOOBTag(t, body, "deck-list")
}

// Deleting a card the caller owns empties the row: htmx swaps the empty body
// over the <li>, so the card disappears without a reload.
func TestRemoveFlashcard_OwnedCard(t *testing.T) {
	user := newTestUser(t, "user@example.com", "correct-horse-battery", true)
	cardID := newTestUUID()
	q := &stubQuerier{
		deleteFlashcard: func(_ context.Context, arg db.DeleteFlashcardParams) (int64, error) {
			if arg.UserID != user.ID {
				t.Fatal("delete was not scoped to the signed-in user")
			}
			if arg.ID != cardID {
				t.Fatal("delete targeted the wrong card")
			}
			return 1, nil
		},
	}
	h, _ := newTestHandler(q)

	rec := serveDelete(h, user, "/flashcard/"+cardID.String(), cardID.String())

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); body != "" {
		t.Errorf("body = %q, want empty so the row swaps away", body)
	}
}

// A card that doesn't exist — or belongs to someone else, which the scoped query
// makes indistinguishable — must 404 rather than report a phantom success.
func TestRemoveFlashcard_NotOwned(t *testing.T) {
	user := newTestUser(t, "user@example.com", "correct-horse-battery", true)
	cardID := newTestUUID()
	q := &stubQuerier{
		deleteFlashcard: func(context.Context, db.DeleteFlashcardParams) (int64, error) {
			return 0, nil // scoped query matched nothing
		},
	}
	h, _ := newTestHandler(q)

	rec := serveDelete(h, user, "/flashcard/"+cardID.String(), cardID.String())

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// A validation failure re-renders the form only. Sending an out-of-band list
// alongside it would be harmless but pointless; more importantly the card must
// not have been written.
func TestCreateFlashcard_ValidationErrorDoesNotWrite(t *testing.T) {
	user := newTestUser(t, "user@example.com", "correct-horse-battery", true)
	q := &stubQuerier{} // every method panics: none should be called
	h, _ := newTestHandler(q)

	rec := serveAs(h.CreateFlashcard, user, http.MethodPost, "/flashcard", url.Values{
		"front": {"   "}, // trimmed to empty
		"back":  {"to conclude a contract"},
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (htmx only swaps on 2xx)", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "front is required") {
		t.Error("expected the validation message in the re-rendered form")
	}
	if strings.Contains(body, `hx-swap-oob`) {
		t.Error("a failed create should not refresh the list")
	}
}
