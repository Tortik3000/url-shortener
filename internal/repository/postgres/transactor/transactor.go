package transactor

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edkin/url-shortener/internal/repository/sqlc"
)

type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
	Queries(ctx context.Context) *sqlc.Queries
}

type transactor struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
}

func New(db *pgxpool.Pool) *transactor {
	return &transactor{
		db: db,
		q:  sqlc.New(db),
	}
}

func (s *transactor) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	ctx = withTx(ctx, tx)

	if fnErr := fn(ctx); fnErr != nil {
		_ = tx.Rollback(ctx)
		return fnErr
	}

	return tx.Commit(ctx)
}

func (s *transactor) Queries(ctx context.Context) *sqlc.Queries {
	if tx, ok := txFromContext(ctx); ok {
		return s.q.WithTx(tx)
	}
	return s.q
}

type ctxKey struct{}

func withTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, ctxKey{}, tx)
}

func txFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(ctxKey{}).(pgx.Tx)
	return tx, ok
}
