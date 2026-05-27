package link_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	linkhandler "github.com/edkin/url-shortener/internal/handler/link"
	"github.com/edkin/url-shortener/internal/handler/link/mocks"
	"github.com/edkin/url-shortener/internal/models"
	"github.com/edkin/url-shortener/internal/service"
)

const baseURL = "http://localhost:8080"

type httpHandler interface {
	Shorten(w http.ResponseWriter, r *http.Request)
	Resolve(w http.ResponseWriter, r *http.Request)
}

func newHandler(ctrl *gomock.Controller) (httpHandler, *mocks.MocklinkService) {
	svc := mocks.NewMocklinkService(ctrl)
	h := linkhandler.New(svc, baseURL)
	return h, svc
}

func TestHandler_Shorten(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		body       string
		setup      func(svc *mocks.MocklinkService)
		wantStatus int
		check      func(t *testing.T, body []byte)
	}{
		{
			name:       "malformed json",
			body:       `{not json`,
			setup:      func(*mocks.MocklinkService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service ErrInvalidURL → 400",
			body: `{"url":"bad"}`,
			setup: func(svc *mocks.MocklinkService) {
				svc.EXPECT().Shorten(gomock.Any(), "bad").Return(models.Link{}, service.ErrInvalidURL)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service ErrCodeCollision → 503",
			body: `{"url":"https://x.com"}`,
			setup: func(svc *mocks.MocklinkService) {
				svc.EXPECT().Shorten(gomock.Any(), "https://x.com").Return(models.Link{}, service.ErrCodeCollision)
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name: "service generic error → 500",
			body: `{"url":"https://x.com"}`,
			setup: func(svc *mocks.MocklinkService) {
				svc.EXPECT().Shorten(gomock.Any(), "https://x.com").Return(models.Link{}, errors.New("boom"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "ok → 201 with short_url",
			body: `{"url":"https://x.com"}`,
			setup: func(svc *mocks.MocklinkService) {
				svc.EXPECT().Shorten(gomock.Any(), "https://x.com").Return(
					models.Link{ShortCode: "abc1234567", OriginalURL: "https://x.com"}, nil)
			},
			wantStatus: http.StatusCreated,
			check: func(t *testing.T, body []byte) {
				t.Helper()
				var resp struct {
					ShortCode string `json:"short_code"`
					ShortURL  string `json:"short_url"`
				}
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, "abc1234567", resp.ShortCode)
				assert.Equal(t, baseURL+"/abc1234567", resp.ShortURL)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			h, svc := newHandler(ctrl)
			tc.setup(svc)

			req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.Shorten(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.check != nil {
				tc.check(t, rec.Body.Bytes())
			}
		})
	}
}

func TestHandler_Resolve(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		code        string
		setup       func(svc *mocks.MocklinkService)
		wantStatus  int
		check       func(t *testing.T, body []byte)
		checkHeader func(t *testing.T, h http.Header)
	}{
		{
			name: "invalid code → 400",
			code: "short",
			setup: func(svc *mocks.MocklinkService) {
				svc.EXPECT().Resolve(gomock.Any(), "short").Return(models.Link{}, service.ErrInvalidCode)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "not found → 404",
			code: "missing001",
			setup: func(svc *mocks.MocklinkService) {
				svc.EXPECT().Resolve(gomock.Any(), "missing001").Return(models.Link{}, service.ErrLinkNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "ok → 302 redirect to original url",
			code: "abc1234567",
			setup: func(svc *mocks.MocklinkService) {
				svc.EXPECT().Resolve(gomock.Any(), "abc1234567").Return(
					models.Link{ShortCode: "abc1234567", OriginalURL: "https://x.com"}, nil)
			},
			wantStatus: http.StatusFound,
			checkHeader: func(t *testing.T, h http.Header) {
				t.Helper()
				assert.Equal(t, "https://x.com", h.Get("Location"))
			},
		},
		{
			name: "service error → 500",
			code: "abc1234567",
			setup: func(svc *mocks.MocklinkService) {
				svc.EXPECT().Resolve(gomock.Any(), "abc1234567").Return(models.Link{}, errors.New("boom"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			h, svc := newHandler(ctrl)
			tc.setup(svc)

			req := httptest.NewRequest(http.MethodGet, "/"+tc.code, nil).
				WithContext(context.Background())
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("code", tc.code)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rec := httptest.NewRecorder()
			h.Resolve(rec, req)

			assert.Equal(t, tc.wantStatus, rec.Code)
			if tc.check != nil {
				tc.check(t, rec.Body.Bytes())
			}
			if tc.checkHeader != nil {
				tc.checkHeader(t, rec.Header())
			}
		})
	}
}
