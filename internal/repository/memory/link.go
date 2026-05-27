package memory

import (
	"context"
	"sync"

	"github.com/edkin/url-shortener/internal/models"
	"github.com/edkin/url-shortener/internal/repository/repoerr"
)

type Repository interface {
	Create(ctx context.Context, link models.Link) (models.Link, error)
	GetByCode(ctx context.Context, code string) (models.Link, error)
	GetByURL(ctx context.Context, url string) (models.Link, error)
}

type linkRepo struct {
	mu     sync.RWMutex
	byCode map[string]models.Link
	byURL  map[string]string // url → code
}

var _ Repository = (*linkRepo)(nil)

func New() *linkRepo {
	return &linkRepo{
		byCode: make(map[string]models.Link),
		byURL:  make(map[string]string),
	}
}

func (r *linkRepo) Create(_ context.Context, link models.Link) (models.Link, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byCode[link.ShortCode]; exists {
		return models.Link{}, repoerr.ErrConflict
	}
	if _, exists := r.byURL[link.OriginalURL]; exists {
		return models.Link{}, repoerr.ErrConflict
	}

	r.byCode[link.ShortCode] = link
	r.byURL[link.OriginalURL] = link.ShortCode
	return link, nil
}

func (r *linkRepo) GetByCode(_ context.Context, code string) (models.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	link, ok := r.byCode[code]
	if !ok {
		return models.Link{}, repoerr.ErrNotFound
	}
	return link, nil
}

func (r *linkRepo) GetByURL(_ context.Context, url string) (models.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	code, ok := r.byURL[url]
	if !ok {
		return models.Link{}, repoerr.ErrNotFound
	}
	return r.byCode[code], nil
}
