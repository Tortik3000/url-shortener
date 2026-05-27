package random

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandomGenerator_FormatAndAlphabet(t *testing.T) {
	t.Parallel()

	g := New()

	const samples = 1_000
	for i := 0; i < samples; i++ {
		code, err := g.Generate()
		require.NoError(t, err)
		require.Len(t, code, codeLength)
		require.True(t, isValidCode(code), "code %q not in alphabet", code)
		for _, r := range code {
			assert.True(t, strings.ContainsRune(alphabet, r), "char %q not in alphabet", r)
		}
	}
}

func TestRandomGenerator_Uniqueness(t *testing.T) {
	t.Parallel()

	g := New()

	const samples = 10_000
	seen := make(map[string]struct{}, samples)
	for i := 0; i < samples; i++ {
		code, err := g.Generate()
		require.NoError(t, err)
		seen[code] = struct{}{}
	}
	require.GreaterOrEqual(t, len(seen), samples-1,
		"duplicate codes generated unexpectedly often — distribution bug?")
}

func TestRandomGenerator_ConcurrentSafe(t *testing.T) {
	t.Parallel()

	g := New()
	const (
		workers   = 8
		perWorker = 1_000
		total     = workers * perWorker
	)

	var (
		mu   sync.Mutex
		seen = make(map[string]struct{}, total)
	)

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				code, err := g.Generate()
				assert.NoError(t, err)
				mu.Lock()
				seen[code] = struct{}{}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	require.GreaterOrEqual(t, len(seen), total-1)
}

func isValidCode(s string) bool {
	if len(s) != codeLength {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !isAlphabetByte(s[i]) {
			return false
		}
	}
	return true
}

func isAlphabetByte(b byte) bool {
	switch {
	case b >= 'a' && b <= 'z':
		return true
	case b >= 'A' && b <= 'Z':
		return true
	case b >= '0' && b <= '9':
		return true
	case b == '_':
		return true
	}
	return false
}
