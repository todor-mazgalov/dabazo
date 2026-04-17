// Package cli implements the uninstall command for dabazo.
package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/todor-mazgalov/dabazo/internal/engines"
	"github.com/todor-mazgalov/dabazo/internal/pkgmgr"
	"github.com/todor-mazgalov/dabazo/internal/registry"
)

var flagPurge bool

// newUninstallCommand creates the uninstall command descriptor.
func newUninstallCommand() *command {
	return &command{
		name:  "uninstall",
		use:   "uninstall",
		short: "Uninstall a database instance and remove it from the registry",
		long: `Stop the instance (if running), uninstall the packages via the same package
manager (with confirmation), and remove the registry entry. Does not delete
the data directory unless --purge is passed.`,
		example: `  dabazo uninstall --name dev
  dabazo uninstall --name dev --purge -y`,
		run: runUninstall,
		localFlags: func(fs *flag.FlagSet) {
			fs.BoolVar(&flagPurge, "purge", false, "also delete the data directory")
		},
	}
}

func runUninstall(args []string) error {
	inst, err := registry.Resolve(flagName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitNotFound)
	}
	printInstanceName(inst.Name)

	if inst.PackageManager == "external" {
		fmt.Fprintln(os.Stderr, "error: instance was added via `registry add`; dabazo does not manage its lifecycle")
		os.Exit(ExitUsage)
	}

	pm, err := pkgmgr.ByName(inst.PackageManager)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	eng, err := resolveEngine(inst.Engine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	plan, err := eng.UninstallPlan(*inst, pm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	printPlan("uninstall", plan, inst.Name)
	confirmOrAbort(flagYes)

	runner := newRunner()
	stopIfRunning(eng, *inst, runner)
	executeUninstallPlan(plan, runner)
	purgeDataIfRequested(inst)

	if err := registry.Remove(inst.Name); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	fmt.Printf("Instance %q uninstalled and removed from registry.\n", inst.Name)
	return nil
}

// stopIfRunning stops the instance if the engine reports it as running.
func stopIfRunning(eng engines.Engine, inst engines.Instance, runner engines.CommandRunner) {
	if eng.IsRunning(inst) {
		if err := eng.Stop(inst, runner); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to stop instance: %v\n", err)
		}
	}
}

// executeUninstallPlan runs each uninstall command from the plan in sequence.
func executeUninstallPlan(plan engines.InstallPlan, runner engines.CommandRunner) {
	for _, c := range plan.Commands {
		if len(c) == 0 {
			continue
		}
		if _, err := runner.Run(c[0], c[1:]...); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(ExitPkgManager)
		}
	}
}

// purgeDataIfRequested removes the data directory when the --purge flag is set.
func purgeDataIfRequested(inst *engines.Instance) {
	if !flagPurge || inst.DataDir == "" {
		return
	}
	if err := os.RemoveAll(inst.DataDir); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to remove data directory: %v\n", err)
		return
	}
	fmt.Printf("Data directory %s removed.\n", inst.DataDir)
}
