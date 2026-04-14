// Package cli implements the start command for dabazo.
package cli

import (
	"fmt"
	"os"

	"dabazo/internal/pkgmgr"
	"dabazo/internal/registry"
)

// newStartCommand creates the start command descriptor.
func newStartCommand() *command {
	return &command{
		name:  "start",
		use:   "start",
		short: "Start a database instance",
		long:  "Start the database service for a registered instance. Idempotent: starting an already-running instance prints \"already running\" and exits 0.",
		example: `  dabazo start
  dabazo start --name dev`,
		run: runStart,
	}
}

func runStart(args []string) error {
	inst, err := registry.Resolve(flagName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitNotFound)
	}

	if inst.PackageManager == "external" {
		fmt.Fprintln(os.Stderr, "error: instance was added via `registry add`; dabazo does not manage its lifecycle")
		os.Exit(ExitUsage)
	}

	if _, err := pkgmgr.ByName(inst.PackageManager); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	eng, err := resolveEngine(inst.Engine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	runner := newRunner()
	if err := eng.Start(*inst, runner); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitPkgManager)
	}

	return nil
}
