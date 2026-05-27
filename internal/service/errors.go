package service

import "errors"

var (
	ErrLinkNotFound  = errors.New("link not found")
	ErrInvalidURL    = errors.New("invalid url")
	ErrInvalidCode   = errors.New("invalid short code")
	ErrCodeCollision = errors.New("could not generate unique code")
)
