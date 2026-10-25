package httpapp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	authhttp "sso/internal/http/auth"
)

type App struct {
	log        *slog.Logger
	httpServer *http.Server
	port       int
}

func New(
	log *slog.Logger,
	authService authhttp.Auth,
	port int,
) *App {
	handler := authhttp.New(authService)

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		log:        log,
		httpServer: httpServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "httpapp.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	log.Info("starting HTTP server", slog.String("addr", a.httpServer.Addr))

	if err := a.httpServer.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "httpapp.Stop"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a.log.With(slog.String("op", op)).
		Info("stopping HTTP server", slog.Int("port", a.port))

	if err := a.httpServer.Shutdown(ctx); err != nil {
		a.log.Error("failed to gracefully stop HTTP server", slog.String("error", err.Error()))
	}
}
