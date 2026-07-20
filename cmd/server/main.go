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
	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	err := godotenv.Load()
	if err != nil {
		logger.Error("error loading .env file", "error", err)
		os.Exit(1)
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

	// Serve cookies with the Secure attribute whenever the site is behind HTTPS.
	// Keep it false for local plain-HTTP development.
	secureCookies := boolFromEnv("SECURE_COOKIES", false, logger)

	sessionManager := scs.New()
	sessionManager.Store = pgxstore.New(pool)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = secureCookies

	queries := db.New(pool)
	h := handlers.New(queries, logger, sessionManager)

	srv := &http.Server{
		Addr:         addr,
		Handler:      h.Routes(sessionManager, secureCookies),
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

func boolFromEnv(key string, def bool, logger *slog.Logger) bool {
	s := os.Getenv(key)
	if s == "" {
		return def
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		logger.Error("invalid boolean env var, using default", "key", key, "value", s, "default", def)
		return def
	}
	return v
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
