package link

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/edkin/url-shortener/internal/handler/helpers"
	"github.com/edkin/url-shortener/internal/models"
	"github.com/edkin/url-shortener/internal/service"
	"github.com/edkin/url-shortener/pkg/logger"
)

//go:generate mockgen -source=handler.go -destination=mocks/handler_mocks.go -package=mocks

type (
	linkService interface {
		Shorten(ctx context.Context, originalURL string) (models.Link, error)
		Resolve(ctx context.Context, code string) (models.Link, error)
	}
)

type handler struct {
	svc     linkService
	baseURL string
}

func New(svc linkService, baseURL string) *handler {
	return &handler{
		svc:     svc,
		baseURL: baseURL,
	}
}

const maxRequestBodyBytes = 1 << 20

func (h *handler) Shorten(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes))
	if err := dec.Decode(&req); err != nil {
		helpers.WriteJSON(r.Context(), w, http.StatusBadRequest, helpers.ErrResp("INVALID_REQUEST", "malformed json body"))
		return
	}

	link, err := h.svc.Shorten(r.Context(), req.URL)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidURL):
			helpers.WriteJSON(r.Context(), w, http.StatusBadRequest, helpers.ErrResp("INVALID_URL", err.Error()))
		case errors.Is(err, service.ErrCodeCollision):
			helpers.WriteJSON(r.Context(), w, http.StatusServiceUnavailable, helpers.ErrResp("RETRY_LATER", err.Error()))
		default:
			logger.FromContext(r.Context()).Error("shorten failed", logger.Error(err))
			helpers.WriteJSON(r.Context(), w, http.StatusInternalServerError, helpers.InternalErr())
		}
		return
	}

	helpers.WriteJSON(r.Context(), w, http.StatusCreated, shortenResponse{
		ShortCode: link.ShortCode,
		ShortURL:  buildShortURL(h.baseURL, link.ShortCode),
	})
}

func (h *handler) Resolve(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	link, err := h.svc.Resolve(r.Context(), code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCode):
			helpers.WriteJSON(r.Context(), w, http.StatusBadRequest, helpers.ErrResp("INVALID_CODE", err.Error()))
		case errors.Is(err, service.ErrLinkNotFound):
			helpers.WriteJSON(r.Context(), w, http.StatusNotFound, helpers.ErrResp("NOT_FOUND", err.Error()))
		default:
			logger.FromContext(r.Context()).Error("resolve failed", logger.Error(err))
			helpers.WriteJSON(r.Context(), w, http.StatusInternalServerError, helpers.InternalErr())
		}
		return
	}

	http.Redirect(w, r, link.OriginalURL, http.StatusFound)
}
