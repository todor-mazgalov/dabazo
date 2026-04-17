// Package cli provides interactive prompting for missing required flags.
package cli

import (
	"fmt"
	"io"
	"strconv"

	"github.com/todor-mazgalov/dabazo/internal/prompt"
)

// requiredFlag describes a flag that can be prompted for in interactive mode.
type requiredFlag struct {
	// name is the long flag name (e.g. "engine").
	name string
	// description is the human-readable prompt text.
	description string
	// defaultValue is shown in the prompt and used when the user presses Enter.
	defaultValue string
	// isMissing returns true when the flag was not provided on the command line.
	isMissing func() bool
	// set applies the user-provided value. For int flags it should parse and
	// validate the input.
	set func(string) error
}

// promptMissing iterates the required flags and prompts the user for any that
// are missing. It uses ReadLineWithDefault when a default is available and
// ReadLine otherwise.
func promptMissing(flags []requiredFlag, r io.Reader, w io.Writer) error {
	for _, rf := range flags {
		if !rf.isMissing() {
			continue
		}
		val, err := prompt.ReadLineWithDefault(rf.description, rf.defaultValue, r, w)
		if err != nil {
			return fmt.Errorf("prompting for --%s: %w", rf.name, err)
		}
		if val == "" {
			return fmt.Errorf("--%s is required", rf.name)
		}
		if err := rf.set(val); err != nil {
			return fmt.Errorf("invalid value for --%s: %w", rf.name, err)
		}
	}
	return nil
}

// stringFlagSetter returns a set function that assigns the value to the
// target string pointer.
func stringFlagSetter(target *string) func(string) error {
	return func(val string) error {
		*target = val
		return nil
	}
}

// intFlagSetter returns a set function that parses val as an integer and
// assigns it to the target.
func intFlagSetter(target *int) func(string) error {
	return func(val string) error {
		n, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("%q is not a valid integer", val)
		}
		*target = n
		return nil
	}
}
