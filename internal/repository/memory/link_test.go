package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edkin/url-shortener/internal/models"
	"github.com/edkin/url-shortener/internal/repository/repoerr"
)

func TestLinkRepo_Create(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		setup   func(r *linkRepo)
		input   models.Link
		wantErr error
	}{
		{
			name:  "ok",
			input: models.Link{ShortCode: "abc1234567", OriginalURL: "https://a.com", CreatedAt: time.Now().UTC()},
		},
		{
			name: "duplicate code",
			setup: func(r *linkRepo) {
				_, _ = r.Create(context.Background(), models.Link{ShortCode: "dup1234567", OriginalURL: "https://other.com"})
			},
			input:   models.Link{ShortCode: "dup1234567", OriginalURL: "https://b.com"},
			wantErr: repoerr.ErrConflict,
		},
		{
			name: "duplicate url",
			setup: func(r *linkRepo) {
				_, _ = r.Create(context.Background(), models.Link{ShortCode: "code000001", OriginalURL: "https://dup.com"})
			},
			input:   models.Link{ShortCode: "code000002", OriginalURL: "https://dup.com"},
			wantErr: repoerr.ErrConflict,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := New()
			if tc.setup != nil {
				tc.setup(repo)
			}

			link, err := repo.Create(context.Background(), tc.input)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.input.ShortCode, link.ShortCode)
			assert.Equal(t, tc.input.OriginalURL, link.OriginalURL)
		})
	}
}

func TestLinkRepo_GetByCode(t *testing.T) {
	t.Parallel()

	repo := New()
	link := models.Link{ShortCode: "code123456", OriginalURL: "https://x.com", CreatedAt: time.Now().UTC()}
	_, err := repo.Create(context.Background(), link)
	require.NoError(t, err)

	t.Run("found", func(t *testing.T) {
		got, err := repo.GetByCode(context.Background(), "code123456")
		require.NoError(t, err)
		assert.Equal(t, link.OriginalURL, got.OriginalURL)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetByCode(context.Background(), "nonexistent")
		assert.ErrorIs(t, err, repoerr.ErrNotFound)
	})
}

func TestLinkRepo_GetByURL(t *testing.T) {
	t.Parallel()

	repo := New()
	link := models.Link{ShortCode: "code123456", OriginalURL: "https://x.com", CreatedAt: time.Now().UTC()}
	_, err := repo.Create(context.Background(), link)
	require.NoError(t, err)

	t.Run("found", func(t *testing.T) {
		got, err := repo.GetByURL(context.Background(), "https://x.com")
		require.NoError(t, err)
		assert.Equal(t, "code123456", got.ShortCode)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetByURL(context.Background(), "https://nope.com")
		assert.ErrorIs(t, err, repoerr.ErrNotFound)
	})
}
