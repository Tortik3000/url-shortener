package random

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	codeLength = 10
	alphabet   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
)

type randomGenerator struct{}

func New() *randomGenerator {
	return &randomGenerator{}
}

func (g *randomGenerator) Generate() (string, error) {
	out := make([]byte, codeLength)
	limit := big.NewInt(int64(len(alphabet)))
	for i := range out {
		n, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", fmt.Errorf("rand.Int: %w", err)
		}
		out[i] = alphabet[n.Int64()]
	}
	return string(out), nil
}
