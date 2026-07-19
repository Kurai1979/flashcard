package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	"github.com/Kurai1979/flashcard/internal/auth"
	"github.com/Kurai1979/flashcard/internal/db"
	"github.com/Kurai1979/flashcard/internal/templates"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (h *Handler) NewUser(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, http.StatusOK, templates.CreateUserPage())
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	// htmx only swaps content on 2xx responses, so validation errors return 200
	// with the re-rendered form rather than a 4xx the client would ignore.
	formError := func(msg string) {
		h.render(w, r, http.StatusOK, templates.CreateUser(templates.CreateUserForm{
			Email: email,
			Error: msg,
		}))
	}

	if err := validatePassword(password); err != nil {
		formError(err.Error())
		return
	}

	e, err := mail.ParseAddress(email)
	if err != nil {
		formError("email is not valid")
		return
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(e.Address))

	_, err = h.Queries.GetUserByEmail(r.Context(), normalizedEmail)
	switch {
	case err == nil:
		formError("email already registered")
		return
	case !errors.Is(err, pgx.ErrNoRows):
		h.serverError(w, r, "lookup user", err)
		return
	}

	hash, err := auth.GenerateHash(password, auth.DefaultParams)
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
			formError("email already registered")
			return
		}
		h.serverError(w, r, "insert user", err)
		return
	}

	h.Logger.InfoContext(r.Context(), "created user", "email", createdUser.Email)
	h.render(w, r, http.StatusCreated, templates.CreateUserSuccess(createdUser.Email))
}

func validatePassword(password string) error {
	const minLength = 12

	if len(password) < minLength {
		return fmt.Errorf("password length must be at least %d", minLength)
	}
	if len(password) > auth.MaxPasswordLength {
		return fmt.Errorf("password must not exceed %d", auth.MaxPasswordLength)
	}

	return nil
}
