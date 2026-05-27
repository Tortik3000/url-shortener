package link

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildShortURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		baseURL string
		code    string
		want    string
	}{
		{
			name:    "no trailing slash",
			baseURL: "http://localhost:8080",
			code:    "abc1234567",
			want:    "http://localhost:8080/abc1234567",
		},
		{
			name:    "single trailing slash",
			baseURL: "http://localhost:8080/",
			code:    "abc1234567",
			want:    "http://localhost:8080/abc1234567",
		},
		{
			name:    "multiple trailing slashes",
			baseURL: "http://localhost:8080///",
			code:    "abc1234567",
			want:    "http://localhost:8080/abc1234567",
		},
		{
			name:    "https with port",
			baseURL: "https://short.example.com:443",
			code:    "xYz_aB9999",
			want:    "https://short.example.com:443/xYz_aB9999",
		},
		{
			name:    "base with path prefix no trailing slash",
			baseURL: "https://example.com/r",
			code:    "abc1234567",
			want:    "https://example.com/r/abc1234567",
		},
		{
			name:    "base with path prefix and trailing slash",
			baseURL: "https://example.com/r/",
			code:    "abc1234567",
			want:    "https://example.com/r/abc1234567",
		},
		{
			name:    "empty base URL",
			baseURL: "",
			code:    "abc1234567",
			want:    "/abc1234567",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, buildShortURL(tc.baseURL, tc.code))
		})
	}
}
