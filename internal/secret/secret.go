// Package secret provides secure random password generation.
package secret

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// base62 is the character set for generated passwords: alphanumeric, no special chars.
const base62 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// GeneratePassword returns a cryptographically random base62 password of the given length.
func GeneratePassword(length int) (string, error) {
	buf := make([]byte, length)
	max := big.NewInt(int64(len(base62)))
	for i := range buf {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("generating random byte: %w", err)
		}
		buf[i] = base62[n.Int64()]
	}
	return string(buf), nil
}
