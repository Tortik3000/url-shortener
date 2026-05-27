package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	migrations "github.com/edkin/url-shortener"
	"github.com/edkin/url-shortener/internal/config"
	linkhandler "github.com/edkin/url-shortener/internal/handler/link"
	"github.com/edkin/url-shortener/internal/middleware"
	"github.com/edkin/url-shortener/internal/models"
	memoryrepo "github.com/edkin/url-shortener/internal/repository/memory"
	postgresrepo "github.com/edkin/url-shortener/internal/repository/postgres/link"
	"github.com/edkin/url-shortener/internal/repository/postgres/transactor"
	linksvc "github.com/edkin/url-shortener/internal/service/link"
	"github.com/edkin/url-shortener/internal/service/link/generator/random"
	"github.com/edkin/url-shortener/pkg/logger"
)

const (
	initialRetryTime = 1 * time.Second
	maxRetryTime     = 5 * time.Second
	maxRetryAttempts = 5
)

type (
	linkRepository interface {
		Create(ctx context.Context, link models.Link) (models.Link, error)
		GetByCode(ctx context.Context, code string) (models.Link, error)
		GetByURL(ctx context.Context, url string) (models.Link, error)
	}
)

func provideRepository(ctx context.Context, logs logger.Logger, cfg *config.Config) (linkRepository, func(), error) {
	switch cfg.Storage.Type {
	case config.StorageMemory:
		logs.Info("using in-memory storage")
		return memoryrepo.New(), func() {}, nil
	case config.StoragePostgres:
		return providePostgresRepository(ctx, logs, cfg)
	default:
		return nil, func() {}, fmt.Errorf("unknown storage type %q", cfg.Storage.Type)
	}
}

func providePostgresRepository(ctx context.Context, logs logger.Logger, cfg *config.Config) (linkRepository, func(), error) {
	pool, err := provideDBPool(ctx, logs, cfg)
	if err != nil {
		return nil, func() {}, err
	}
	if err := runMigrations(logs, pool); err != nil {
		pool.Close()
		return nil, func() {}, err
	}
	return postgresrepo.New(transactor.New(pool)), pool.Close, nil
}

func provideDBPool(ctx context.Context, logs logger.Logger, cfg *config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}

	if err := pingWithRetry(ctx, pool, logs); err != nil {
		pool.Close()
		return nil, fmt.Errorf("provideDBPool: %w", err)
	}

	logs.Info("database connection pool established")
	return pool, nil
}

func pingWithRetry(ctx context.Context, pool *pgxpool.Pool, logs logger.Logger) error {
	delay := initialRetryTime
	var err error
	for attempt := 0; attempt < maxRetryAttempts; attempt++ {
		if err = ctx.Err(); err != nil {
			return err
		}

		if err = pool.Ping(ctx); err == nil {
			return nil
		}

		logs.Warn(
			"database ping failed",
			logger.NewField("attempt", attempt+1),
			logger.Error(err),
		)

		if attempt == maxRetryAttempts-1 {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		delay = min(delay*2, maxRetryTime)
	}
	return err
}

func runMigrations(logs logger.Logger, pool *pgxpool.Pool) error {
	goose.SetBaseFS(migrations.Migrations)

	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose.SetDialect: %w", err)
	}

	if err := goose.Up(sqlDB, "migrations"); err != nil {
		return fmt.Errorf("goose.Up: %w", err)
	}
	logs.Info("migrations applied")
	return nil
}

func provideHTTPServer(logs logger.Logger, cfg *config.Config, repo linkRepository) *http.Server {
	router := provideRouter(logs, cfg, repo)

	return &http.Server{
		Handler:      router,
		Addr:         net.JoinHostPort("", cfg.Server.Port),
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
	}
}

func provideRouter(logs logger.Logger, cfg *config.Config, repo linkRepository) *chi.Mux {
	gen := random.New()
	svc := linksvc.New(repo, gen)
	linkH := linkhandler.New(svc, cfg.BaseURL)

	r := chi.NewRouter()
	r.Use(middleware.WithLogger(logs))

	r.Post("/shorten", linkH.Shorten)
	r.Get("/{code}", linkH.Resolve)

	return r
}
