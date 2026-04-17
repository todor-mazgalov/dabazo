// Package cli provides a lightweight command dispatcher built on the Go standard
// library. It replaces the cobra dependency with flag.FlagSet-based parsing,
// subcommand routing, and auto-generated help text.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// command describes a single CLI command or subcommand group.
type command struct {
	// name is the bare command name used in dispatch (e.g. "install", "user").
	name string
	// use is the one-line usage synopsis shown in help (e.g. "migrate <filepath>").
	use string
	// short is a brief description for the parent's subcommand listing.
	short string
	// long is an extended description shown in the command's own help.
	long string
	// example contains indented example invocations.
	example string
	// run is the handler invoked after flag parsing. Nil for group commands.
	run func(args []string) error
	// localFlags registers command-specific flags on the given flag set.
	localFlags func(fs *flag.FlagSet)
	// subcommands holds child commands for group commands.
	subcommands []*command
	// requiredFlags lists flags that can be prompted in interactive mode.
	requiredFlags []requiredFlag
}

// dispatch resolves and executes the appropriate command from the argument list.
// It handles subcommand lookup, flag merging, help detection, and argument
// validation before delegating to the command's run function.
func dispatch(root *command, osArgs []string) error {
	args := osArgs[1:] // skip program name
	return resolveAndRun(root, args, "dabazo")
}

// resolveAndRun walks the subcommand tree, parsing flags at each level.
func resolveAndRun(cmd *command, args []string, path string) error {
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		if sub := findSubcommand(cmd, args[0]); sub != nil {
			return resolveAndRun(sub, args[1:], path+" "+sub.name)
		}
	}

	if cmd.run == nil && len(cmd.subcommands) > 0 {
		return handleGroupHelp(cmd, args, path)
	}

	if cmd.run == nil {
		printHelp(os.Stdout, cmd, path)
		return nil
	}

	return parseAndRun(cmd, args, path)
}

// findSubcommand returns the named child command, or nil if not found.
func findSubcommand(parent *command, name string) *command {
	for _, sub := range parent.subcommands {
		if sub.name == name {
			return sub
		}
	}
	return nil
}

// handleGroupHelp prints help or an error for group commands that have no run handler.
func handleGroupHelp(cmd *command, args []string, path string) error {
	if containsHelpFlag(args) || len(args) == 0 {
		printHelp(os.Stdout, cmd, path)
		return nil
	}
	fmt.Fprintf(os.Stderr, "error: unknown subcommand %q for %s\n", args[0], path)
	printHelp(os.Stderr, cmd, path)
	os.Exit(ExitUsage)
	return nil
}

// parseAndRun builds a merged flag set, parses arguments, and calls the handler.
func parseAndRun(cmd *command, args []string, path string) error {
	fs := flag.NewFlagSet(path, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	registerGlobalFlags(fs)
	if cmd.localFlags != nil {
		cmd.localFlags(fs)
	}

	var helpFlag bool
	fs.BoolVar(&helpFlag, "help", false, "show help")
	// Register -h as a help alias only if the command does not already use it.
	if fs.Lookup("h") == nil {
		fs.BoolVar(&helpFlag, "h", false, "short for --help")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printHelp(os.Stdout, cmd, path)
			return nil
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		printUsageLine(os.Stderr, cmd, path)
		os.Exit(ExitUsage)
	}

	if helpFlag {
		printHelp(os.Stdout, cmd, path)
		return nil
	}

	if flagInteractive && len(cmd.requiredFlags) > 0 {
		if err := promptMissing(cmd.requiredFlags, os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(ExitUsage)
		}
	}

	return cmd.run(fs.Args())
}

// containsHelpFlag returns true if args contains -h, --help, or -help.
func containsHelpFlag(args []string) bool {
	for _, a := range args {
		if a == "-h" || a == "--help" || a == "-help" {
			return true
		}
		if a == "--" {
			return false
		}
	}
	return false
}

// registerGlobalFlags adds the persistent (global) flags to a flag set.
func registerGlobalFlags(fs *flag.FlagSet) {
	fs.StringVar(&flagName, "name", "", "logical instance name")
	fs.StringVar(&flagName, "n", "", "short for --name")
	fs.StringVar(&flagEngine, "engine", "", "engine[:version] (e.g. postgres:16)")
	fs.StringVar(&flagEngine, "e", "", "short for --engine")
	fs.IntVar(&flagPort, "port", 0, "TCP port the instance listens on")
	fs.IntVar(&flagPort, "p", 0, "short for --port")
	fs.BoolVar(&flagYes, "yes", false, "skip confirmation prompts")
	fs.BoolVar(&flagYes, "y", false, "short for --yes")
	fs.BoolVar(&flagInteractive, "interactive", false, "prompt for missing required parameters")
	fs.BoolVar(&flagInteractive, "it", false, "short for --interactive")
}

// printHelp writes a formatted help page for the command to the given writer.
func printHelp(w io.Writer, cmd *command, path string) {
	if cmd.name == "dabazo" {
		printMascot(w)
	}
	if cmd.long != "" {
		fmt.Fprintln(w, cmd.long)
	} else if cmd.short != "" {
		fmt.Fprintln(w, cmd.short)
	}
	fmt.Fprintln(w)

	printUsageLine(w, cmd, path)

	if len(cmd.subcommands) > 0 {
		printSubcommandList(w, cmd)
	}

	if cmd.run != nil {
		printFlagHelp(w, cmd)
	}

	if cmd.example != "" {
		fmt.Fprintln(w, "Examples:")
		fmt.Fprintln(w, cmd.example)
		fmt.Fprintln(w)
	}
}

// printUsageLine writes the "Usage:" line appropriate for the command type.
func printUsageLine(w io.Writer, cmd *command, path string) {
	if cmd.use != "" {
		fmt.Fprintf(w, "Usage:\n  %s %s\n", path, usageSuffix(cmd))
	} else if len(cmd.subcommands) > 0 {
		fmt.Fprintf(w, "Usage:\n  %s [command]\n", path)
	}
	fmt.Fprintln(w)
}

// usageSuffix returns the portion of the usage line after the command path.
// For group commands it returns "[command]"; for leaf commands it strips the
// command name prefix from the use field (e.g. "user <username>" -> "<username>")
// and appends "[flags]".
func usageSuffix(cmd *command) string {
	if len(cmd.subcommands) > 0 {
		return "[command]"
	}
	suffix := cmd.use
	if idx := strings.Index(suffix, " "); idx >= 0 {
		suffix = suffix[idx+1:]
	} else {
		suffix = ""
	}
	if suffix != "" {
		return suffix + " [flags]"
	}
	return "[flags]"
}

// printSubcommandList writes the available subcommands section.
func printSubcommandList(w io.Writer, cmd *command) {
	fmt.Fprintln(w, "Available Commands:")
	for _, sub := range cmd.subcommands {
		fmt.Fprintf(w, "  %-15s %s\n", sub.name, sub.short)
	}
	fmt.Fprintln(w)
}

// printFlagHelp writes the flags section by building a temporary flag set.
func printFlagHelp(w io.Writer, cmd *command) {
	fs := flag.NewFlagSet("help", flag.ContinueOnError)
	fs.SetOutput(w)
	registerGlobalFlags(fs)
	if cmd.localFlags != nil {
		cmd.localFlags(fs)
	}

	fmt.Fprintln(w, "Flags:")
	fs.PrintDefaults()
	fmt.Fprintln(w)
}
