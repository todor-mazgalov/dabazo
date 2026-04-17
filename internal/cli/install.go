// Package cli implements the install command for dabazo.
package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/todor-mazgalov/dabazo/internal/engines"
	"github.com/todor-mazgalov/dabazo/internal/pkgmgr"
	"github.com/todor-mazgalov/dabazo/internal/registry"
)

// newInstallCommand creates the install command descriptor.
func newInstallCommand() *command {
	return &command{
		name:  "install",
		use:   "install",
		short: "Install and register a new database instance",
		long: `Install and register a new database instance via the native package manager.

Prints the exact commands it will run and prompts for confirmation before executing.
The instance is left stopped after installation; run 'dabazo start' next.`,
		example: `  dabazo install --engine postgres:16 --port 5432 --name dev
  dabazo install -e postgres:17 -p 5433 -n next -y`,
		run: runInstall,
		requiredFlags: []requiredFlag{
			{
				name:        "engine",
				description: "Engine[:version]",
				isMissing:   func() bool { return flagEngine == "" },
				set:         stringFlagSetter(&flagEngine),
			},
			{
				name:        "port",
				description: "TCP port",
				isMissing:   func() bool { return flagPort == 0 },
				set:         intFlagSetter(&flagPort),
			},
			{
				name:        "name",
				description: "Instance name",
				isMissing:   func() bool { return flagName == "" },
				set:         stringFlagSetter(&flagName),
			},
		},
	}
}

func runInstall(args []string) error {
	if flagEngine == "" {
		fmt.Fprintln(os.Stderr, "error: --engine is required for install")
		os.Exit(ExitUsage)
	}
	if flagPort == 0 {
		fmt.Fprintln(os.Stderr, "error: --port is required for install")
		os.Exit(ExitUsage)
	}
	if flagName == "" {
		fmt.Fprintln(os.Stderr, "error: --name is required for install")
		os.Exit(ExitUsage)
	}

	existing, err := registry.Find(flagName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}
	if existing != nil {
		fmt.Fprintf(os.Stderr, "error: instance %q already exists in registry\n", flagName)
		os.Exit(ExitAlreadyExists)
	}

	engineName, version := parseDB(flagEngine)
	eng, err := resolveEngine(engineName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitUsage)
	}

	pm, err := pkgmgr.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitPkgManager)
	}

	if version == "" {
		version = "16"
	}

	plan, err := eng.Plan(version, flagPort, pm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	printPlan("install", plan, flagName)
	confirmOrAbort(flagYes)

	runner := newRunner()
	if err := eng.Install(plan, runner); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitPkgManager)
	}

	inst := engines.Instance{
		Name:           flagName,
		Engine:         engineName,
		Version:        version,
		Port:           flagPort,
		Host:           "localhost",
		InstalledAt:    time.Now().UTC().Format(time.RFC3339),
		PackageManager: pm.Name(),
		ServiceName:    plan.ServiceName,
		DataDir:        plan.DataDir,
		BinDir:         plan.BinDir,
	}
	if err := registry.Add(inst); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	fmt.Printf("Instance %q registered successfully.\n", flagName)
	return nil
}
