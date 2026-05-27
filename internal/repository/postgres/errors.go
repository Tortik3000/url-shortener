package postgres

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/edkin/url-shortener/internal/repository/repoerr"
)

const pgUniqueViolation = "23505"

func MapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return repoerr.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return repoerr.ErrConflict
	}
	return err
}

func QueryError(query string, err error) error {
	return fmt.Errorf("queries.%s: %w", query, MapError(err))
}
