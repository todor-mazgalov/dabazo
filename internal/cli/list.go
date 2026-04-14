// Package cli implements the list command for dabazo.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"dabazo/internal/engines"
	"dabazo/internal/engines/postgres"
	"dabazo/internal/registry"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List all registered database instances",
		Long:    "Print the registry: one line per instance with name, engine:version, port, and running/stopped status.",
		Example: "  dabazo list",
		RunE:    runList,
	}
}

func runList(cmd *cobra.Command, args []string) error {
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
