-- name: CreateLink :one
INSERT INTO links (short_code, original_url, created_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreateLinkIfNotExists :one
INSERT INTO links (short_code, original_url, created_at)
VALUES ($1, $2, $3)
ON CONFLICT (original_url) DO NOTHING
RETURNING *;

-- name: GetLinkByCode :one
SELECT * FROM links WHERE short_code = $1;

-- name: GetLinkByURL :one
SELECT * FROM links WHERE original_url = $1;
