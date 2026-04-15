// Package cli implements the list command for dabazo.
package cli

import (
	"fmt"
	"os"

	"github.com/todor-mazgalov/dabazo/internal/engines"
	"github.com/todor-mazgalov/dabazo/internal/engines/postgres"
	"github.com/todor-mazgalov/dabazo/internal/registry"
)

// newListCommand creates the list command descriptor.
func newListCommand() *command {
	return &command{
		name:    "list",
		use:     "list",
		short:   "List all registered database instances",
		long:    "Print the registry: one line per instance with name, engine:version, port, and running/stopped status.",
		example: "  dabazo list",
		run:     runList,
	}
}

func runList(args []string) error {
	instances, err := registry.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(ExitGeneric)
	}

	if len(instances) == 0 {
		fmt.Println("No instances registered.")
		return nil
	}

	fmt.Printf("%-15s %-15s %-8s %s\n", "NAME", "ENGINE", "PORT", "STATUS")
	for _, inst := range instances {
		status := checkStatus(inst.Engine, inst)
		fmt.Printf("%-15s %-15s %-8d %s\n",
			inst.Name,
			inst.Engine+":"+inst.Version,
			inst.Port,
			status,
		)
	}
	return nil
}

// checkStatus probes whether an instance is running using the engine driver.
func checkStatus(engineName string, inst engines.Instance) string {
	switch engineName {
	case "postgres":
		drv := &postgres.Driver{}
		if drv.IsRunning(inst) {
			return "running"
		}
		return "stopped"
	default:
		return "unknown"
	}
}
