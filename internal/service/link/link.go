package link

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/edkin/url-shortener/internal/models"
	"github.com/edkin/url-shortener/internal/repository/repoerr"
	"github.com/edkin/url-shortener/internal/service"
)

//go:generate mockgen -source=link.go -destination=mocks/link_mocks.go -package=mocks

const maxCollisionRetries = 5

type (
	linkRepository interface {
		Create(ctx context.Context, link models.Link) (models.Link, error)
		GetByCode(ctx context.Context, code string) (models.Link, error)
		GetByURL(ctx context.Context, url string) (models.Link, error)
	}

	codeGenerator interface {
		Generate() (string, error)
	}

	clock interface {
		Now() time.Time
	}
)

type Service interface {
	Shorten(ctx context.Context, originalURL string) (models.Link, error)
	Resolve(ctx context.Context, code string) (models.Link, error)
}

type linkService struct {
	repo  linkRepository
	gen   codeGenerator
	clock clock
}

var _ Service = (*linkService)(nil)

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

func New(repo linkRepository, gen codeGenerator) *linkService {
	return &linkService{
		repo:  repo,
		gen:   gen,
		clock: realClock{},
	}
}

func newWithClock(repo linkRepository, gen codeGenerator, c clock) *linkService {
	return &linkService{
		repo:  repo,
		gen:   gen,
		clock: c,
	}
}

func (s *linkService) Shorten(ctx context.Context, originalURL string) (models.Link, error) {
	if err := validateURL(originalURL); err != nil {
		return models.Link{}, err
	}

	existing, err := s.repo.GetByURL(ctx, originalURL)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, repoerr.ErrNotFound) {
		return models.Link{}, fmt.Errorf("repo.GetByURL: %w", err)
	}

	for attempt := 0; attempt < maxCollisionRetries; attempt++ {
		code, genErr := s.gen.Generate()
		if genErr != nil {
			return models.Link{}, fmt.Errorf("generator.Generate: %w", genErr)
		}

		link := models.Link{
			ShortCode:   code,
			OriginalURL: originalURL,
			CreatedAt:   s.clock.Now(),
		}
		created, createErr := s.repo.Create(ctx, link)
		if createErr == nil {
			return created, nil
		}
		if !errors.Is(createErr, repoerr.ErrConflict) {
			return models.Link{}, fmt.Errorf("repo.Create: %w", createErr)
		}

		if existing, fetchErr := s.repo.GetByURL(ctx, originalURL); fetchErr == nil {
			return existing, nil
		}
	}

	return models.Link{}, service.ErrCodeCollision
}

func (s *linkService) Resolve(ctx context.Context, code string) (models.Link, error) {
	link, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repoerr.ErrNotFound) {
			return models.Link{}, service.ErrLinkNotFound
		}
		return models.Link{}, fmt.Errorf("repo.GetByCode: %w", err)
	}
	return link, nil
}

func validateURL(raw string) error {
	if raw == "" {
		return service.ErrInvalidURL
	}
	u, err := url.Parse(raw)
	if err != nil {
		return service.ErrInvalidURL
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return service.ErrInvalidURL
	}
	if u.Host == "" {
		return service.ErrInvalidURL
	}
	return nil
}
