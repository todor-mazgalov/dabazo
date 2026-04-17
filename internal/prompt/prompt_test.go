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

func TestReadLineWithDefault_AcceptsDefault(t *testing.T) {
	r := strings.NewReader("\n")
	w := &bytes.Buffer{}
	got, err := ReadLineWithDefault("Port", "5432", r, w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "5432" {
		t.Errorf("expected '5432', got %q", got)
	}
	if !strings.Contains(w.String(), "[5432]") {
		t.Errorf("prompt should show default value, got %q", w.String())
	}
}

func TestReadLineWithDefault_OverridesDefault(t *testing.T) {
	r := strings.NewReader("9999\n")
	w := &bytes.Buffer{}
	got, err := ReadLineWithDefault("Port", "5432", r, w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "9999" {
		t.Errorf("expected '9999', got %q", got)
	}
}

func TestReadLineWithDefault_EmptyDefaultBehavesLikeReadLine(t *testing.T) {
	r := strings.NewReader("hello\n")
	w := &bytes.Buffer{}
	got, err := ReadLineWithDefault("Value", "", r, w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
}

func TestReadLineWithDefault_EmptyDefaultNoInput(t *testing.T) {
	r := strings.NewReader("")
	w := &bytes.Buffer{}
	_, err := ReadLineWithDefault("Value", "", r, w)
	if err == nil {
		t.Error("expected error on empty input with no default")
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

// --------------------------------------------------------------------------
// Additional ReadLineWithDefault tests
// --------------------------------------------------------------------------

func TestReadLineWithDefault_TrimWhitespace(t *testing.T) {
	r := strings.NewReader("  hello  \n")
	w := &bytes.Buffer{}
	got, err := ReadLineWithDefault("Value", "", r, w)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
}

func TestReadLineWithDefault_PromptFormatWithDefault(t *testing.T) {
	r := strings.NewReader("\n")
	w := &bytes.Buffer{}
	_, _ = ReadLineWithDefault("Port", "5432", r, w)
	prompt := w.String()
	if prompt != "Port [5432]: " {
		t.Errorf("prompt format = %q, want %q", prompt, "Port [5432]: ")
	}
}

func TestReadLineWithDefault_PromptFormatWithoutDefault(t *testing.T) {
	r := strings.NewReader("val\n")
	w := &bytes.Buffer{}
	_, _ = ReadLineWithDefault("Name", "", r, w)
	prompt := w.String()
	if prompt != "Name: " {
		t.Errorf("prompt format = %q, want %q", prompt, "Name: ")
	}
}

func TestReadLine_WritesPrompt(t *testing.T) {
	r := strings.NewReader("val\n")
	w := &bytes.Buffer{}
	_, _ = ReadLine("Enter value: ", r, w)
	if w.String() != "Enter value: " {
		t.Errorf("prompt output = %q, want %q", w.String(), "Enter value: ")
	}
}
