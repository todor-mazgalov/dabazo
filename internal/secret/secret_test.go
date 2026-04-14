// Package secret tests for random password generation.
package secret

import (
	"strings"
	"testing"
)

func TestGeneratePassword_Length(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"short", 8},
		{"standard", 32},
		{"long", 64},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pw, err := GeneratePassword(tt.length)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(pw) != tt.length {
				t.Errorf("expected length %d, got %d", tt.length, len(pw))
			}
		})
	}
}

func TestGeneratePassword_Base62Only(t *testing.T) {
	pw, err := GeneratePassword(1000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, ch := range pw {
		if !strings.ContainsRune(base62, ch) {
			t.Errorf("password contains non-base62 character: %c", ch)
		}
	}
}

func TestGeneratePassword_Unique(t *testing.T) {
	pw1, _ := GeneratePassword(32)
	pw2, _ := GeneratePassword(32)
	if pw1 == pw2 {
		t.Error("two consecutive passwords should not be identical")
	}
}
