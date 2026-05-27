package link

import (
	"context"

	"github.com/edkin/url-shortener/internal/models"
	"github.com/edkin/url-shortener/internal/repository/postgres"
	"github.com/edkin/url-shortener/internal/repository/sqlc"
)

type (
	txManager interface {
		Queries(ctx context.Context) *sqlc.Queries
	}
)

type Repository interface {
	Create(ctx context.Context, link models.Link) (models.Link, error)
	GetByCode(ctx context.Context, code string) (models.Link, error)
	GetByURL(ctx context.Context, url string) (models.Link, error)
}

type linkRepo struct {
	tx txManager
}

var _ Repository = (*linkRepo)(nil)

func New(tx txManager) *linkRepo {
	return &linkRepo{
		tx: tx,
	}
}

func (r *linkRepo) Create(ctx context.Context, link models.Link) (models.Link, error) {
	row, err := r.tx.Queries(ctx).CreateLink(ctx, sqlc.CreateLinkParams{
		ShortCode:   link.ShortCode,
		OriginalUrl: link.OriginalURL,
		CreatedAt:   timeToPgTs(link.CreatedAt),
	})
	if err != nil {
		return models.Link{}, postgres.QueryError("CreateLink", err)
	}
	return toModel(row), nil
}

func (r *linkRepo) GetByCode(ctx context.Context, code string) (models.Link, error) {
	row, err := r.tx.Queries(ctx).GetLinkByCode(ctx, code)
	if err != nil {
		return models.Link{}, postgres.QueryError("GetLinkByCode", err)
	}
	return toModel(row), nil
}

func (r *linkRepo) GetByURL(ctx context.Context, url string) (models.Link, error) {
	row, err := r.tx.Queries(ctx).GetLinkByURL(ctx, url)
	if err != nil {
		return models.Link{}, postgres.QueryError("GetLinkByURL", err)
	}
	return toModel(row), nil
}
