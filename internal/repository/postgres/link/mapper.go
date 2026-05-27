package link

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/edkin/url-shortener/internal/models"
	"github.com/edkin/url-shortener/internal/repository/sqlc"
)

func timeToPgTs(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}

func pgTsToTime(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time.UTC()
}

func toModel(l sqlc.Link) models.Link {
	return models.Link{
		ShortCode:   l.ShortCode,
		OriginalURL: l.OriginalUrl,
		CreatedAt:   pgTsToTime(l.CreatedAt),
	}
}
