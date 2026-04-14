// Package prompt tests for confirmation and input utilities.
package prompt

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfirm_Yes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"lowercase y", "y\n", true},
		{"uppercase Y", "Y\n", true},
		{"no", "n\n", false},
		{"empty", "\n", false},
		{"random text", "maybe\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			w := &bytes.Buffer{}
			got := Confirm("Proceed? ", false, r, w)
			if got != tt.want {
				t.Errorf("Confirm(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestConfirm_SkipWithYesFlag(t *testing.T) {
	r := strings.NewReader("")
	w := &bytes.Buffer{}
	got := Confirm("Proceed? ", true, r, w)
	if !got {
		t.Error("expected true when yes=true")
	}
}

func TestReadLine(t *testing.T) {
	r := strings.NewReader("hello world\n")
	w := &bytes.Buffer{}
	got, err := ReadLine("Enter: ", r, w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello world" {
		t.Errorf("expected 'hello world', got %q", got)
	}
}

func TestReadLine_Empty(t *testing.T) {
	r := strings.NewReader("")
	w := &bytes.Buffer{}
	_, err := ReadLine("Enter: ", r, w)
	if err == nil {
		t.Error("expected error on empty input")
	}
}
