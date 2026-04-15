// Package cli helper functions shared across command implementations.
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/todor-mazgalov/dabazo/internal/engines"
	"github.com/todor-mazgalov/dabazo/internal/engines/postgres"
	"github.com/todor-mazgalov/dabazo/internal/executor"
	"github.com/todor-mazgalov/dabazo/internal/prompt"
)

// parseDB parses a --db flag value into engine name and version.
func parseDB(db string) (string, string) {
	parts := strings.SplitN(db, ":", 2)
	engine := parts[0]
	version := ""
	if len(parts) == 2 {
		version = parts[1]
	}
	return engine, version
}

// resolveEngine returns the engine driver for the given engine name.
func resolveEngine(name string) (engines.Engine, error) {
	switch name {
	case "postgres":
		return &postgres.Driver{}, nil
	default:
		return nil, fmt.Errorf("unsupported engine: %q (supported: postgres)", name)
	}
}

// newRunner creates the default command runner.
func newRunner() engines.CommandRunner {
	return &executor.OSRunner{}
}

// printPlan prints the install/uninstall plan for user review.
func printPlan(action string, plan engines.InstallPlan, instanceName string) {
	fmt.Printf("dabazo will run the following to %s %s:%s on port %d:\n\n",
		action, plan.EngineName, plan.Version, plan.Port)
	fmt.Printf("  Package manager : %s\n", plan.PkgManager)
	fmt.Printf("  Source          : %s\n", plan.Source)
	fmt.Printf("  Packages        : %s\n", strings.Join(plan.Packages, ", "))
	fmt.Println("  Commands:")
	for _, cmd := range plan.Commands {
		fmt.Printf("    %s\n", strings.Join(cmd, " "))
	}
	if len(plan.PostInstall) > 0 {
		fmt.Println("  Post-install:")
		for _, step := range plan.PostInstall {
			fmt.Printf("    %s\n", step)
		}
	}
	fmt.Printf("    listening port:        %d\n", plan.Port)
	fmt.Printf("    registered as:         %s\n\n", instanceName)
}

// confirmOrAbort prompts for confirmation and exits if the user declines.
func confirmOrAbort(yes bool) {
	ok := prompt.Confirm("Proceed? [y/N] ", yes, os.Stdin, os.Stdout)
	if !ok {
		fmt.Fprintln(os.Stderr, "aborted by user")
		os.Exit(ExitAborted)
	}
}
