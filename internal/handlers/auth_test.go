package handlers

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"uuid"

	"github.com/Kurai1979/flashcard/internal/auth"
	"github.com/Kurai1979/flashcard/internal/db"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// stubQuerier implements db.Querier with per-method hooks; unset methods panic
// so a test only wires what it exercises.
type stubQuerier struct {
	getUserByEmail  func(ctx context.Context, email string) (db.User, error)
	updateLastLogin func(ctx context.Context, id pgtype.UUID) error
	createDeck      func(ctx context.Context, arg db.CreateDeckParams) (db.Deck, error)
	listDecks       func(ctx context.Context, id pgtype.UUID) ([]db.Deck, error)
	createFlashcard func(ctx context.Context, arg db.CreateFlashcardParams) (db.Flashcard, error)
	listFlashcards  func(ctx context.Context, arg db.ListFlashcardsParams) ([]db.ListFlashcardsRow, error)
	deleteFlashcard func(ctx context.Context, arg db.DeleteFlashcardParams) (int64, error)
}

func (s *stubQuerier) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return s.getUserByEmail(ctx, email)
}

func (s *stubQuerier) UpdateLastLogin(ctx context.Context, id pgtype.UUID) error {
	if s.updateLastLogin == nil {
		return nil
	}
	return s.updateLastLogin(ctx, id)
}

func (s *stubQuerier) CreateUser(context.Context, db.CreateUserParams) (db.User, error) {
	panic("CreateUser not implemented")
}
func (s *stubQuerier) GetUserById(context.Context, pgtype.UUID) (db.User, error) {
	panic("GetUserById not implemented")
}
func (s *stubQuerier) HealthCheck(context.Context) (int32, error) {
	panic("HealthCheck not implemented")
}

func (s *stubQuerier) CreateDeck(ctx context.Context, arg db.CreateDeckParams) (db.Deck, error) {
	if s.createDeck == nil {
		panic("CreateDeck not implemented")
	}
	return s.createDeck(ctx, arg)
}
func (s *stubQuerier) GetDeck(context.Context, db.GetDeckParams) (db.Deck, error) {
	panic("GetDeck not implemented")
}
func (s *stubQuerier) ListDecks(ctx context.Context, id pgtype.UUID) ([]db.Deck, error) {
	if s.listDecks == nil {
		panic("ListDecks not implemented")
	}
	return s.listDecks(ctx, id)
}
func (s *stubQuerier) UpdateDeck(context.Context, db.UpdateDeckParams) (db.Deck, error) {
	panic("UpdateDeck not implemented")
}
func (s *stubQuerier) DeleteDeck(context.Context, db.DeleteDeckParams) (int64, error) {
	panic("DeleteDeck not implemented")
}
func (s *stubQuerier) AddFlashcardToDeck(context.Context, db.AddFlashcardToDeckParams) (int64, error) {
	panic("AddFlashcardToDeck not implemented")
}
func (s *stubQuerier) RemoveFlashcardFromDeck(context.Context, db.RemoveFlashcardFromDeckParams) (int64, error) {
	panic("RemoveFlashcardFromDeck not implemented")
}

func (s *stubQuerier) CreateFlashcard(ctx context.Context, arg db.CreateFlashcardParams) (db.Flashcard, error) {
	if s.createFlashcard == nil {
		panic("CreateFlashcard not implemented")
	}
	return s.createFlashcard(ctx, arg)
}
func (s *stubQuerier) UpdateFlashcard(context.Context, db.UpdateFlashcardParams) (db.Flashcard, error) {
	panic("UpdateFlashcard not implemented")
}
func (s *stubQuerier) DeleteFlashcard(ctx context.Context, arg db.DeleteFlashcardParams) (int64, error) {
	if s.deleteFlashcard == nil {
		panic("DeleteFlashcard not implemented")
	}
	return s.deleteFlashcard(ctx, arg)
}
func (s *stubQuerier) ListFlashcards(ctx context.Context, arg db.ListFlashcardsParams) ([]db.ListFlashcardsRow, error) {
	if s.listFlashcards == nil {
		panic("ListFlashcards not implemented")
	}
	return s.listFlashcards(ctx, arg)
}

func newTestUser(t *testing.T, email, password string, active bool) db.User {
	t.Helper()
	hash, err := auth.GenerateHash(password, auth.DefaultParams)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	return db.User{
		ID:           pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Email:        email,
		PasswordHash: hash,
		IsActive:     active,
	}
}

// serveLogin runs a POST /login request through the scs LoadAndSave middleware
// so the handler sees a live session context.
func serveLogin(h *Handler, sm *scs.SessionManager, form url.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	sm.LoadAndSave(http.HandlerFunc(h.Login)).ServeHTTP(rec, req)
	return rec
}

func newTestHandler(q db.Querier) (*Handler, *scs.SessionManager) {
	sm := scs.New() // default in-memory store
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(q, logger, sm), sm
}

func TestLogin_Success(t *testing.T) {
	user := newTestUser(t, "user@example.com", "correct-horse-battery", true)
	var updated bool
	q := &stubQuerier{
		getUserByEmail: func(_ context.Context, email string) (db.User, error) {
			if email != "user@example.com" {
				t.Fatalf("unexpected email lookup: %q", email)
			}
			return user, nil
		},
		updateLastLogin: func(_ context.Context, id pgtype.UUID) error {
			updated = true
			return nil
		},
	}
	h, sm := newTestHandler(q)

	rec := serveLogin(h, sm, url.Values{
		"email":    {"user@example.com"},
		"password": {"correct-horse-battery"},
	})

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("HX-Redirect"); got != "/dashboard" {
		t.Fatalf("HX-Redirect = %q, want /dashboard", got)
	}
	if rec.Header().Get("Set-Cookie") == "" {
		t.Fatalf("expected a session cookie to be set")
	}
	if !updated {
		t.Fatalf("expected UpdateLastLogin to be called")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	user := newTestUser(t, "user@example.com", "correct-horse-battery", true)
	q := &stubQuerier{
		getUserByEmail: func(context.Context, string) (db.User, error) { return user, nil },
	}
	h, sm := newTestHandler(q)

	rec := serveLogin(h, sm, url.Values{
		"email":    {"user@example.com"},
		"password": {"wrong-password-here"},
	})

	assertGenericLoginError(t, rec)
}

func TestLogin_UnknownEmail(t *testing.T) {
	q := &stubQuerier{
		getUserByEmail: func(context.Context, string) (db.User, error) {
			return db.User{}, pgx.ErrNoRows
		},
	}
	h, sm := newTestHandler(q)

	rec := serveLogin(h, sm, url.Values{
		"email":    {"nobody@example.com"},
		"password": {"correct-horse-battery"},
	})

	assertGenericLoginError(t, rec)
}

func TestLogin_InactiveUser(t *testing.T) {
	user := newTestUser(t, "user@example.com", "correct-horse-battery", false)
	q := &stubQuerier{
		getUserByEmail: func(context.Context, string) (db.User, error) { return user, nil },
	}
	h, sm := newTestHandler(q)

	rec := serveLogin(h, sm, url.Values{
		"email":    {"user@example.com"},
		"password": {"correct-horse-battery"},
	})

	assertGenericLoginError(t, rec)
}

// assertGenericLoginError checks a failed login re-renders the form (200) with
// the generic message and never establishes a session.
func assertGenericLoginError(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "invalid email or password") {
		t.Fatalf("body missing generic error, got: %s", rec.Body.String())
	}
	if rec.Header().Get("HX-Redirect") != "" {
		t.Fatalf("unexpected HX-Redirect on failed login")
	}
	if rec.Header().Get("Set-Cookie") != "" {
		t.Fatalf("no session cookie should be set on failed login")
	}
}
