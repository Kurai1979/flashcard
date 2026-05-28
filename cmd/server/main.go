package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Kurai1979/flashcard/internal/db"
	"github.com/Kurai1979/flashcard/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	err := godotenv.Load()
	if err != nil {
		logger.Error("error loading .env file", "error", err)
	}

	port := portFromEnv(8080, logger)
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

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	shutdownCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("starting server", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to start serving", "error", err)
			os.Exit(1)
		}
	}()

	<-shutdownCtx.Done()
	stop()
	logger.Info("shutting down server")

	shutdownTimeout, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownTimeout); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info("server stopped")
}

func portFromEnv(def int, logger *slog.Logger) int {
	s := os.Getenv("PORT")
	if s == "" {
		return def
	}
	p, err := strconv.Atoi(s)
	if err != nil || p < 1024 || p > 65535 {
		logger.Error("invalid PORT, must be 1023-65535", "port", s, "default", def)
		return def
	}
	return p
}
