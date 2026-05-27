package link

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/edkin/url-shortener/internal/models"
	"github.com/edkin/url-shortener/internal/repository/repoerr"
	"github.com/edkin/url-shortener/internal/service"
	"github.com/edkin/url-shortener/internal/service/link/mocks"
)

type fakeClock struct{ t time.Time }

func (c fakeClock) Now() time.Time { return c.t }

func newService(t *testing.T, ctrl *gomock.Controller) (*linkService, *mocks.MocklinkRepository, *mocks.MockcodeGenerator) {
	t.Helper()
	repo := mocks.NewMocklinkRepository(ctrl)
	gen := mocks.NewMockcodeGenerator(ctrl)
	now := time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC)
	return newWithClock(repo, gen, fakeClock{t: now}), repo, gen
}

func TestLinkService_Shorten(t *testing.T) {
	t.Parallel()

	const validURL = "https://example.com/path"
	const newCode = "abcDEF1234"

	cases := []struct {
		name    string
		url     string
		setup   func(repo *mocks.MocklinkRepository, gen *mocks.MockcodeGenerator)
		wantErr error
		check   func(t *testing.T, link models.Link)
	}{
		{
			name:    "invalid url empty",
			url:     "",
			setup:   func(*mocks.MocklinkRepository, *mocks.MockcodeGenerator) {},
			wantErr: service.ErrInvalidURL,
		},
		{
			name:    "invalid url scheme",
			url:     "ftp://x.com",
			setup:   func(*mocks.MocklinkRepository, *mocks.MockcodeGenerator) {},
			wantErr: service.ErrInvalidURL,
		},
		{
			name:    "invalid url no host",
			url:     "http://",
			setup:   func(*mocks.MocklinkRepository, *mocks.MockcodeGenerator) {},
			wantErr: service.ErrInvalidURL,
		},
		{
			name: "existing url returns existing code",
			url:  validURL,
			setup: func(repo *mocks.MocklinkRepository, _ *mocks.MockcodeGenerator) {
				repo.EXPECT().GetByURL(gomock.Any(), validURL).Return(
					models.Link{ShortCode: "existing01", OriginalURL: validURL}, nil)
			},
			check: func(t *testing.T, link models.Link) {
				t.Helper()
				assert.Equal(t, "existing01", link.ShortCode)
			},
		},
		{
			name: "new url created",
			url:  validURL,
			setup: func(repo *mocks.MocklinkRepository, gen *mocks.MockcodeGenerator) {
				repo.EXPECT().GetByURL(gomock.Any(), validURL).Return(models.Link{}, repoerr.ErrNotFound)
				gen.EXPECT().Generate().Return(newCode, nil)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, l models.Link) (models.Link, error) { return l, nil })
			},
			check: func(t *testing.T, link models.Link) {
				t.Helper()
				assert.Equal(t, newCode, link.ShortCode)
				assert.Equal(t, validURL, link.OriginalURL)
			},
		},
		{
			name: "code collision retried once then succeeds",
			url:  validURL,
			setup: func(repo *mocks.MocklinkRepository, gen *mocks.MockcodeGenerator) {
				repo.EXPECT().GetByURL(gomock.Any(), validURL).Return(models.Link{}, repoerr.ErrNotFound)
				gomock.InOrder(
					gen.EXPECT().Generate().Return("code000001", nil),
					gen.EXPECT().Generate().Return("code000002", nil),
				)
				gomock.InOrder(
					repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(models.Link{}, repoerr.ErrConflict),
					repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
						func(_ context.Context, l models.Link) (models.Link, error) { return l, nil }),
				)
				repo.EXPECT().GetByURL(gomock.Any(), validURL).Return(models.Link{}, repoerr.ErrNotFound)
			},
			check: func(t *testing.T, link models.Link) {
				t.Helper()
				assert.Equal(t, "code000002", link.ShortCode)
			},
		},
		{
			name: "conflict resolves to existing url race winner",
			url:  validURL,
			setup: func(repo *mocks.MocklinkRepository, gen *mocks.MockcodeGenerator) {
				gomock.InOrder(
					repo.EXPECT().GetByURL(gomock.Any(), validURL).Return(models.Link{}, repoerr.ErrNotFound),
					repo.EXPECT().GetByURL(gomock.Any(), validURL).Return(
						models.Link{ShortCode: "raceWinner", OriginalURL: validURL}, nil),
				)
				gen.EXPECT().Generate().Return("any1234567", nil)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(models.Link{}, repoerr.ErrConflict)
			},
			check: func(t *testing.T, link models.Link) {
				t.Helper()
				assert.Equal(t, "raceWinner", link.ShortCode)
			},
		},
		{
			name: "all retries collide",
			url:  validURL,
			setup: func(repo *mocks.MocklinkRepository, gen *mocks.MockcodeGenerator) {
				repo.EXPECT().GetByURL(gomock.Any(), validURL).Return(models.Link{}, repoerr.ErrNotFound).Times(1)
				gen.EXPECT().Generate().Return("collision1", nil).Times(maxCollisionRetries)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(models.Link{}, repoerr.ErrConflict).Times(maxCollisionRetries)
				repo.EXPECT().GetByURL(gomock.Any(), validURL).Return(models.Link{}, repoerr.ErrNotFound).Times(maxCollisionRetries)
			},
			wantErr: service.ErrCodeCollision,
		},
		{
			name: "generator error",
			url:  validURL,
			setup: func(repo *mocks.MocklinkRepository, gen *mocks.MockcodeGenerator) {
				repo.EXPECT().GetByURL(gomock.Any(), validURL).Return(models.Link{}, repoerr.ErrNotFound)
				gen.EXPECT().Generate().Return("", errors.New("boom"))
			},
			wantErr: nil,
		},
		{
			name: "repo GetByURL transient error",
			url:  validURL,
			setup: func(repo *mocks.MocklinkRepository, _ *mocks.MockcodeGenerator) {
				repo.EXPECT().GetByURL(gomock.Any(), validURL).Return(models.Link{}, errors.New("db down"))
			},
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			svc, repo, gen := newService(t, ctrl)
			tc.setup(repo, gen)

			link, err := svc.Shorten(context.Background(), tc.url)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			if tc.check == nil {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			tc.check(t, link)
		})
	}
}

func TestLinkService_Resolve(t *testing.T) {
	t.Parallel()

	const code = "validCode1"

	cases := []struct {
		name    string
		code    string
		setup   func(repo *mocks.MocklinkRepository)
		wantErr error
		check   func(t *testing.T, link models.Link)
	}{
		{
			name: "found",
			code: code,
			setup: func(repo *mocks.MocklinkRepository) {
				repo.EXPECT().GetByCode(gomock.Any(), code).Return(
					models.Link{ShortCode: code, OriginalURL: "https://x.com"}, nil)
			},
			check: func(t *testing.T, link models.Link) {
				t.Helper()
				assert.Equal(t, "https://x.com", link.OriginalURL)
			},
		},
		{
			name: "not found",
			code: code,
			setup: func(repo *mocks.MocklinkRepository) {
				repo.EXPECT().GetByCode(gomock.Any(), code).Return(models.Link{}, repoerr.ErrNotFound)
			},
			wantErr: service.ErrLinkNotFound,
		},
		{
			name: "repo error",
			code: code,
			setup: func(repo *mocks.MocklinkRepository) {
				repo.EXPECT().GetByCode(gomock.Any(), code).Return(models.Link{}, errors.New("db down"))
			},
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			svc, repo, _ := newService(t, ctrl)
			tc.setup(repo)

			link, err := svc.Resolve(context.Background(), tc.code)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			if tc.check == nil {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			tc.check(t, link)
		})
	}
}
