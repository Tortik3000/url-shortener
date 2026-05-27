package helpers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrResp(t *testing.T) {
	t.Parallel()

	resp := ErrResp("BAD_REQ", "value")
	b, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.JSONEq(t, `{"error":{"code":"BAD_REQ","message":"value"}}`, string(b))
}

func TestInternalErr(t *testing.T) {
	t.Parallel()

	resp := InternalErr()
	assert.Equal(t, "INTERNAL_ERROR", resp.Error.Code)
	assert.Equal(t, "internal server error", resp.Error.Message)
}

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	type payload struct {
		Foo string `json:"foo"`
		N   int    `json:"n"`
	}

	cases := []struct {
		name       string
		status     int
		body       any
		wantStatus int
		wantBody   string
	}{
		{
			name:       "struct ok",
			status:     http.StatusOK,
			body:       payload{Foo: "bar", N: 42},
			wantStatus: http.StatusOK,
			wantBody:   `{"foo":"bar","n":42}`,
		},
		{
			name:       "map created",
			status:     http.StatusCreated,
			body:       map[string]any{"k": "v"},
			wantStatus: http.StatusCreated,
			wantBody:   `{"k":"v"}`,
		},
		{
			name:       "error response uses ErrResp",
			status:     http.StatusBadRequest,
			body:       ErrResp("BAD", "msg"),
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"BAD","message":"msg"}}`,
		},
		{
			name:       "nil body encodes as null",
			status:     http.StatusOK,
			body:       nil,
			wantStatus: http.StatusOK,
			wantBody:   `null`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			WriteJSON(context.Background(), rec, tc.status, tc.body)

			res := rec.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.wantStatus, res.StatusCode)
			assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
			assert.JSONEq(t, tc.wantBody, rec.Body.String())
		})
	}
}
