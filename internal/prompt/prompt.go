// Package prompt provides confirmation and credential input utilities.
package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// Confirm prints a message and asks for y/N confirmation.
// Returns true if the user types "y" or "Y". Returns false on anything else.
// If yes is true, skips the prompt and returns true immediately.
func Confirm(msg string, yes bool, r io.Reader, w io.Writer) bool {
	fmt.Fprint(w, msg)
	if yes {
		fmt.Fprintln(w)
		return true
	}
	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		answer := strings.TrimSpace(scanner.Text())
		return strings.EqualFold(answer, "y")
	}
	return false
}

// ReadLine reads a single line of input from the given reader, with a prompt.
func ReadLine(prompt string, r io.Reader, w io.Writer) (string, error) {
	fmt.Fprint(w, prompt)
	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}
	return "", fmt.Errorf("no input received")
}

// ReadLineWithDefault reads a single line of input, returning defaultVal when
// the user presses Enter without typing anything. The prompt is displayed as
// "<prompt> [<defaultVal>]: " when a default is provided.
func ReadLineWithDefault(prompt string, defaultVal string, r io.Reader, w io.Writer) (string, error) {
	if defaultVal != "" {
		fmt.Fprintf(w, "%s [%s]: ", prompt, defaultVal)
	} else {
		fmt.Fprintf(w, "%s: ", prompt)
	}
	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		val := strings.TrimSpace(scanner.Text())
		if val == "" {
			return defaultVal, nil
		}
		return val, nil
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}
	if defaultVal != "" {
		return defaultVal, nil
	}
	return "", fmt.Errorf("no input received")
}

// ReadPassword reads a password from the terminal with hidden input.
func ReadPassword(prompt string, w io.Writer) (string, error) {
	fmt.Fprint(w, prompt)
	fd := int(os.Stdin.Fd())
	pw, err := term.ReadPassword(fd)
	fmt.Fprintln(w)
	if err != nil {
		return "", fmt.Errorf("reading password: %w", err)
	}
	return string(pw), nil
}
