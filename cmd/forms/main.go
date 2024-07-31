package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/marcheneli/forms/internal/config"
	"github.com/marcheneli/forms/internal/lib/logger/handlers/slogpretty"

	createSchema "github.com/marcheneli/forms/internal/http-server/handlers/schemas/create"
	deleteSchema "github.com/marcheneli/forms/internal/http-server/handlers/schemas/delete"
	listSchema "github.com/marcheneli/forms/internal/http-server/handlers/schemas/list"
	updateSchema "github.com/marcheneli/forms/internal/http-server/handlers/schemas/update"

	createField "github.com/marcheneli/forms/internal/http-server/handlers/fields/create"
	deleteField "github.com/marcheneli/forms/internal/http-server/handlers/fields/delete"
	listField "github.com/marcheneli/forms/internal/http-server/handlers/fields/list"
	updateField "github.com/marcheneli/forms/internal/http-server/handlers/fields/update"

	mwLogger "github.com/marcheneli/forms/internal/http-server/middleware/logger"
	"github.com/marcheneli/forms/internal/lib/logger/sl"
	fieldsStorage "github.com/marcheneli/forms/internal/storage/sqlite/fields"
	schemasStorage "github.com/marcheneli/forms/internal/storage/sqlite/schemas"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting")

	schemasStorage, err := schemasStorage.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init fields storage", sl.Err(err))
		os.Exit(1)
	}

	fieldsStorage, err := fieldsStorage.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init fields storage", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/schemas", func(r chi.Router) {
		r.Post("/", createSchema.New(log, schemasStorage))
		r.Put("/", updateSchema.New(log, schemasStorage))
		r.Delete("/", deleteSchema.New(log, schemasStorage))
		r.Get("/", listSchema.New(log, schemasStorage))
	})

	router.Route("/fields", func(r chi.Router) {
		r.Post("/", createField.New(log, fieldsStorage))
		r.Put("/", updateField.New(log, fieldsStorage))
		r.Delete("/", deleteField.New(log, fieldsStorage))
		r.Get("/", listField.New(log, fieldsStorage))
	})

	log.Info("starting server", slog.String("address", cfg.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	// TODO: move timeout to config
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	// TODO: close storage

	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default: // If env config is invalid, set prod settings by default due to security
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
