package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

import (
	"context"
	"log/slog"
	"os"

	"github.com/Kurai1979/flashcard/internal/db"
	"github.com/Kurai1979/flashcard/internal/handlers"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	err := godotenv.Load()
	if err != nil {
		logger.Error("error loading .env file", "error", err)
	}

	port := portFromEnv(8080)
	addr := fmt.Sprintf(":%d", port)
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Error("error connecting to database", "error", err)
		os.Exit(1)
	}
	if err := pool.Ping(ctx); err != nil {
		logger.Error("error connecting to database", "pool", err)
		os.Exit(1)
	}

	defer pool.Close()

	queries := db.New(pool)
	h := handlers.New(queries, logger)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Get("/healthz", h.Health)
	r.Post("/users", h.CreateUser)

	logger.Info("starting server", "addr", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Error("failed to start serving", "error", err)
		os.Exit(1)
	}
}

func portFromEnv(def int) int {
	s := os.Getenv("PORT")
	if s == "" {
		return def
	}
	p, err := strconv.Atoi(s)
	if err != nil || p < 1023 || p > 65535 {
		slog.Error("invalid PORT, must be 1023-65535", "port", s, "default", def)
		return def
	}
	return p
}
