package app

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/edkin/url-shortener/internal/config"
	"github.com/edkin/url-shortener/pkg/logger"
)

const (
	gracefulShutdownTimeout = 5 * time.Second
	writeTimeout            = 10 * time.Second
	readTimeout             = 10 * time.Second
)

func Run(logs logger.Logger, cfg *config.Config) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	repo, cleanup, err := provideRepository(ctx, logs, cfg)
	if err != nil {
		logs.Fatal("failed to provide repository", logger.Error(err))
	}
	defer cleanup()

	server := provideHTTPServer(logs, cfg, repo)

	go gracefulShutdown(ctx, server, logs)

	logs.Info("server started",
		logger.NewField("address", server.Addr),
		logger.NewField("storage", cfg.Storage.Type),
	)

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logs.Fatal("server start error", logger.Error(err))
	}
}

func gracefulShutdown(ctx context.Context, srv *http.Server, logs logger.Logger) {
	<-ctx.Done()
	logs.Info("server is shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logs.Error("server shutdown failed", logger.Error(err))
	}
	logs.Info("server stopped")
}
