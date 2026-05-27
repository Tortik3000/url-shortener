package models

import "time"

type Link struct {
	ShortCode   string
	OriginalURL string
	CreatedAt   time.Time
}
