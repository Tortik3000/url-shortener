package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	cases := []struct {
		name    string
		env     map[string]string
		wantErr bool
		check   func(t *testing.T, c *Config)
	}{
		{
			name: "postgres default",
			env: map[string]string{
				"STORAGE":      "postgres",
				"DATABASE_URL": "postgres://localhost/db",
			},
			check: func(t *testing.T, c *Config) {
				t.Helper()
				assert.Equal(t, "8080", c.Server.Port)
				assert.Equal(t, StoragePostgres, c.Storage.Type)
				assert.Equal(t, "postgres://localhost/db", c.Database.URL)
				assert.Equal(t, "http://localhost:8080", c.BaseURL)
			},
		},
		{
			name: "memory storage",
			env: map[string]string{
				"STORAGE":  "memory",
				"PORT":     "9999",
				"BASE_URL": "http://shortener.local",
			},
			check: func(t *testing.T, c *Config) {
				t.Helper()
				assert.Equal(t, "9999", c.Server.Port)
				assert.Equal(t, StorageMemory, c.Storage.Type)
				assert.Empty(t, c.Database.URL)
				assert.Equal(t, "http://shortener.local", c.BaseURL)
			},
		},
		{
			name: "postgres missing DATABASE_URL",
			env: map[string]string{
				"STORAGE": "postgres",
			},
			wantErr: true,
		},
		{
			name: "unknown storage type",
			env: map[string]string{
				"STORAGE": "redis",
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			c, err := New()
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tc.check != nil {
				tc.check(t, c)
			}
		})
	}
}
