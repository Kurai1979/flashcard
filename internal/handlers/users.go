package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Kurai1979/flashcard/internal/auth"
	"github.com/Kurai1979/flashcard/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type createUserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type createUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))

	_, err := h.Queries.GetUserByEmail(r.Context(), normalizedEmail)

	switch {
	case err == nil:
		http.Error(w, "email already registered", http.StatusConflict)
		return
	case !errors.Is(err, pgx.ErrNoRows):
		h.serverError(w, r, "lookup user", err)
		return
	}

	hash, err := auth.GenerateHash(req.Password, auth.DefaultParams)
	if err != nil {
		h.serverError(w, r, "hash password", err)
		return
	}

	newUser := db.CreateUserParams{
		Email:        normalizedEmail,
		PasswordHash: hash,
		IsActive:     true,
		IsVerified:   false,
	}

	createdUser, err := h.Queries.CreateUser(r.Context(), newUser)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			http.Error(w, "email already registered", http.StatusConflict)
			return
		}
		h.serverError(w, r, "insert user", err)
		return
	}

	response := createUserResponse{
		ID:    uuid.UUID(createdUser.ID.Bytes).String(),
		Email: createdUser.Email,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.Logger.ErrorContext(r.Context(), "encode response", "err", err)
		return
	}

	h.Logger.Info("created user", "email", createdUser.Email)
}
