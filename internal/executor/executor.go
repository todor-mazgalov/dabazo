// Package executor provides a concrete command runner using os/exec.
package executor

import (
	"fmt"
	"os"
	"os/exec"
)

// OSRunner executes commands via os/exec.
type OSRunner struct{}

// Run executes a command and returns combined output.
func (r *OSRunner) Run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("command %q failed: %w\n%s", name, err, string(out))
	}
	return out, nil
}

// RunWithEnv executes a command with additional environment variables.
func (r *OSRunner) RunWithEnv(env []string, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = os.Stdin
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("command %q failed: %w\n%s", name, err, string(out))
	}
	return out, nil
}
