// Package cli tests the lightweight command dispatcher in command.go.
package cli

import (
	"bytes"
	"flag"
	"strings"
	"testing"
)

// --------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------

// resetGlobalFlags resets all package-level flag variables to their zero
// values so tests do not bleed state into each other.
func resetGlobalFlags(t *testing.T) {
	t.Helper()
	flagName = ""
	flagEngine = ""
	flagPort = 0
	flagYes = false
	flagHost = ""
	flagUser = ""
	flagDatabase = ""
	flagSchema = ""
	flagForce = false
	flagPurge = false
}

// captureHelp calls printHelp for the given command and returns the output.
func captureHelp(t *testing.T, cmd *command, path string) string {
	t.Helper()
	var buf bytes.Buffer
	printHelp(&buf, cmd, path)
	return buf.String()
}

// --------------------------------------------------------------------------
// findSubcommand
// --------------------------------------------------------------------------

func TestFindSubcommand_Found(t *testing.T) {
	root := newRootCommand()
	sub := findSubcommand(root, "install")
	if sub == nil {
		t.Fatal("expected to find subcommand 'install', got nil")
	}
	if sub.name != "install" {
		t.Errorf("expected name 'install', got %q", sub.name)
	}
}

func TestFindSubcommand_NotFound(t *testing.T) {
	root := newRootCommand()
	sub := findSubcommand(root, "badcmd")
	if sub != nil {
		t.Errorf("expected nil for unknown subcommand, got %+v", sub)
	}
}

func TestFindSubcommand_TwoLevel_Config(t *testing.T) {
	root := newRootCommand()
	configCmd := findSubcommand(root, "create")
	if configCmd == nil {
		t.Fatal("expected to find 'create' group command")
	}
	userCmd := findSubcommand(configCmd, "user")
	if userCmd == nil {
		t.Fatal("expected to find 'config user' subcommand")
	}
	if userCmd.name != "user" {
		t.Errorf("expected name 'user', got %q", userCmd.name)
	}
}

func TestFindSubcommand_TwoLevel_Registry(t *testing.T) {
	root := newRootCommand()
	regCmd := findSubcommand(root, "registry")
	if regCmd == nil {
		t.Fatal("expected to find 'registry' group command")
	}
	for _, sub := range []string{"add", "remove"} {
		found := findSubcommand(regCmd, sub)
		if found == nil {
			t.Errorf("expected to find 'registry %s', got nil", sub)
		}
	}
}

// --------------------------------------------------------------------------
// containsHelpFlag
// --------------------------------------------------------------------------

func TestContainsHelpFlag(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want bool
	}{
		{"long --help", []string{"--help"}, true},
		{"short -h", []string{"-h"}, true},
		{"single-dash -help", []string{"-help"}, true},
		{"mixed with other flags", []string{"--name", "dev", "--help"}, true},
		{"no help flag", []string{"--name", "dev", "--engine", "postgres:16"}, false},
		{"empty args", []string{}, false},
		{"help after separator", []string{"--", "--help"}, false},
		{"help before separator", []string{"--help", "--"}, true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := containsHelpFlag(tc.args)
			if got != tc.want {
				t.Errorf("containsHelpFlag(%v) = %v, want %v", tc.args, got, tc.want)
			}
		})
	}
}

// --------------------------------------------------------------------------
// usageSuffix
// --------------------------------------------------------------------------

func TestUsageSuffix_LeafWithPositionalArg(t *testing.T) {
	cmd := &command{use: "user <username>"}
	got := usageSuffix(cmd)
	if got != "<username> [flags]" {
		t.Errorf("usageSuffix = %q, want %q", got, "<username> [flags]")
	}
}

func TestUsageSuffix_LeafNoPositionalArg(t *testing.T) {
	cmd := &command{use: "install"}
	got := usageSuffix(cmd)
	if got != "[flags]" {
		t.Errorf("usageSuffix = %q, want %q", got, "[flags]")
	}
}

func TestUsageSuffix_GroupCommand(t *testing.T) {
	cmd := &command{
		use: "create",
		subcommands: []*command{
			{name: "user"},
		},
	}
	got := usageSuffix(cmd)
	if got != "[command]" {
		t.Errorf("usageSuffix = %q, want %q", got, "[command]")
	}
}

func TestUsageSuffix_TwoPositionalArgs(t *testing.T) {
	cmd := &command{use: "snapshot <db> <path>"}
	got := usageSuffix(cmd)
	if got != "<db> <path> [flags]" {
		t.Errorf("usageSuffix = %q, want %q", got, "<db> <path> [flags]")
	}
}

// --------------------------------------------------------------------------
// printHelp — content verification
// --------------------------------------------------------------------------

func TestPrintHelp_RootContainsAllSubcommands(t *testing.T) {
	root := newRootCommand()
	out := captureHelp(t, root, "dabazo")

	expected := []string{
		"install", "start", "stop", "create", "migrate",
		"snapshot", "registry", "list", "uninstall",
	}
	for _, name := range expected {
		if !strings.Contains(out, name) {
			t.Errorf("root help missing subcommand %q", name)
		}
	}
}

func TestPrintHelp_RootUsageLine(t *testing.T) {
	root := newRootCommand()
	out := captureHelp(t, root, "dabazo")
	if !strings.Contains(out, "Usage:") {
		t.Error("root help missing 'Usage:' section")
	}
	if !strings.Contains(out, "dabazo [command]") {
		t.Error("root help missing 'dabazo [command]' usage line")
	}
}

func TestPrintHelp_RootLongDescription(t *testing.T) {
	root := newRootCommand()
	out := captureHelp(t, root, "dabazo")
	if !strings.Contains(out, "cross-platform") {
		t.Error("root help missing long description text")
	}
}

func TestPrintHelp_InstallContainsFlagsSection(t *testing.T) {
	cmd := newInstallCommand()
	out := captureHelp(t, cmd, "dabazo install")
	if !strings.Contains(out, "Flags:") {
		t.Error("install help missing 'Flags:' section")
	}
}

func TestPrintHelp_InstallContainsGlobalFlags(t *testing.T) {
	cmd := newInstallCommand()
	out := captureHelp(t, cmd, "dabazo install")
	for _, flag := range []string{"-name", "-engine", "-port", "-yes", "-y"} {
		if !strings.Contains(out, flag) {
			t.Errorf("install help missing global flag %q", flag)
		}
	}
}

func TestPrintHelp_InstallContainsExamples(t *testing.T) {
	cmd := newInstallCommand()
	out := captureHelp(t, cmd, "dabazo install")
	if !strings.Contains(out, "Examples:") {
		t.Error("install help missing 'Examples:' section")
	}
}

func TestPrintHelp_ConfigGroupShowsSubcommandList(t *testing.T) {
	cmd := newCreateCommand()
	out := captureHelp(t, cmd, "dabazo create")
	if !strings.Contains(out, "user") {
		t.Error("config help missing 'user' subcommand listing")
	}
	if !strings.Contains(out, "Available Commands:") {
		t.Error("config help missing 'Available Commands:' section")
	}
}

func TestPrintHelp_ConfigGroupNoFlagsSection(t *testing.T) {
	cmd := newCreateCommand()
	out := captureHelp(t, cmd, "dabazo create")
	// Group commands have no run handler so printHelp does not emit a Flags section.
	if strings.Contains(out, "Flags:") {
		t.Error("config group help should NOT contain a 'Flags:' section (no run handler)")
	}
}

func TestPrintHelp_ConfigUserUsageLine(t *testing.T) {
	cmd := newCreateUserCommand()
	out := captureHelp(t, cmd, "dabazo config user")
	if !strings.Contains(out, "<username>") {
		t.Error("config user help missing '<username>' in usage line")
	}
}

func TestPrintHelp_RegistryGroupListsAddRemove(t *testing.T) {
	cmd := newRegistryCommand()
	out := captureHelp(t, cmd, "dabazo registry")
	for _, sub := range []string{"add", "remove"} {
		if !strings.Contains(out, sub) {
			t.Errorf("registry help missing subcommand %q", sub)
		}
	}
}

func TestPrintHelp_RegistryAddContainsHostFlag(t *testing.T) {
	cmd := newRegistryAddCommand()
	out := captureHelp(t, cmd, "dabazo registry add")
	if !strings.Contains(out, "-host") {
		t.Error("registry add help missing '-host' local flag")
	}
}

func TestPrintHelp_SnapshotContainsForceFlag(t *testing.T) {
	cmd := newSnapshotCommand()
	out := captureHelp(t, cmd, "dabazo snapshot")
	if !strings.Contains(out, "-force") {
		t.Error("snapshot help missing '-force' local flag")
	}
}

func TestPrintHelp_MigrateContainsUserFlag(t *testing.T) {
	cmd := newMigrateCommand()
	out := captureHelp(t, cmd, "dabazo migrate")
	if !strings.Contains(out, "-user") {
		t.Error("migrate help missing '-user' local flag")
	}
}

func TestPrintHelp_UninstallContainsPurgeFlag(t *testing.T) {
	cmd := newUninstallCommand()
	out := captureHelp(t, cmd, "dabazo uninstall")
	if !strings.Contains(out, "-purge") {
		t.Error("uninstall help missing '-purge' local flag")
	}
}

// --------------------------------------------------------------------------
// registerGlobalFlags — flag set registration
// --------------------------------------------------------------------------

func TestRegisterGlobalFlags_AllFlagsPresent(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	registerGlobalFlags(fs)

	for _, name := range []string{"name", "n", "engine", "e", "port", "p", "yes", "y"} {
		if fs.Lookup(name) == nil {
			t.Errorf("expected global flag %q to be registered, but it was not", name)
		}
	}
}

func TestRegisterGlobalFlags_ParseName(t *testing.T) {
	resetGlobalFlags(t)
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	registerGlobalFlags(fs)
	if err := fs.Parse([]string{"--name", "myinstance"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if flagName != "myinstance" {
		t.Errorf("flagName = %q, want %q", flagName, "myinstance")
	}
}

func TestRegisterGlobalFlags_ParseEngine(t *testing.T) {
	resetGlobalFlags(t)
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	registerGlobalFlags(fs)
	if err := fs.Parse([]string{"--engine", "postgres:16"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if flagEngine != "postgres:16" {
		t.Errorf("flagEngine = %q, want %q", flagEngine, "postgres:16")
	}
}

func TestRegisterGlobalFlags_ParsePort(t *testing.T) {
	resetGlobalFlags(t)
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	registerGlobalFlags(fs)
	if err := fs.Parse([]string{"--port", "5432"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if flagPort != 5432 {
		t.Errorf("flagPort = %d, want 5432", flagPort)
	}
}

func TestRegisterGlobalFlags_ParseYesLong(t *testing.T) {
	resetGlobalFlags(t)
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	registerGlobalFlags(fs)
	if err := fs.Parse([]string{"--yes"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if !flagYes {
		t.Error("flagYes should be true after --yes")
	}
}

func TestRegisterGlobalFlags_ParseYesShort(t *testing.T) {
	resetGlobalFlags(t)
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	registerGlobalFlags(fs)
	if err := fs.Parse([]string{"-y"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if !flagYes {
		t.Error("flagYes should be true after -y")
	}
}

func TestRegisterGlobalFlags_UnknownFlag(t *testing.T) {
	resetGlobalFlags(t)
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(new(bytes.Buffer)) // suppress output
	registerGlobalFlags(fs)
	err := fs.Parse([]string{"--nonexistent"})
	if err == nil {
		t.Error("expected error for unknown flag, got nil")
	}
}

// --------------------------------------------------------------------------
// parseAndRun — flag merging, help detection, positional args
// --------------------------------------------------------------------------

func TestParseAndRun_HelpFlagPreventsRun(t *testing.T) {
	resetGlobalFlags(t)
	var ran bool
	cmd := &command{
		name: "test",
		use:  "test",
		run: func(args []string) error {
			ran = true
			return nil
		},
	}
	// parseAndRun should detect --help and print help without calling run.
	err := parseAndRun(cmd, []string{"--help"}, "dabazo test")
	if err != nil {
		t.Errorf("parseAndRun returned error on --help: %v", err)
	}
	if ran {
		t.Error("run handler should NOT be called when --help is present")
	}
}

func TestParseAndRun_ShortHelpFlagPreventsRun(t *testing.T) {
	resetGlobalFlags(t)
	var ran bool
	cmd := &command{
		name: "test",
		use:  "test",
		run: func(args []string) error {
			ran = true
			return nil
		},
	}
	err := parseAndRun(cmd, []string{"-h"}, "dabazo test")
	if err != nil {
		t.Errorf("parseAndRun returned error on -h: %v", err)
	}
	if ran {
		t.Error("run handler should NOT be called when -h is present")
	}
}

func TestParseAndRun_GlobalFlagsParsedBeforeRun(t *testing.T) {
	resetGlobalFlags(t)
	var capturedArgs []string
	cmd := &command{
		name: "test",
		use:  "test",
		run: func(args []string) error {
			capturedArgs = args
			return nil
		},
	}
	err := parseAndRun(cmd, []string{"--name", "dev", "--port", "5432"}, "dabazo test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flagName != "dev" {
		t.Errorf("flagName = %q, want %q", flagName, "dev")
	}
	if flagPort != 5432 {
		t.Errorf("flagPort = %d, want 5432", flagPort)
	}
	if len(capturedArgs) != 0 {
		t.Errorf("expected no positional args, got %v", capturedArgs)
	}
}

func TestParseAndRun_LocalFlagMergedWithGlobal(t *testing.T) {
	resetGlobalFlags(t)
	var ran bool
	cmd := &command{
		name: "test",
		use:  "test",
		run: func(args []string) error {
			ran = true
			return nil
		},
		localFlags: func(fs *flag.FlagSet) {
			fs.StringVar(&flagHost, "host", "localhost", "host address")
		},
	}
	err := parseAndRun(cmd, []string{"--host", "10.0.0.1", "--name", "x"}, "dabazo test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ran {
		t.Error("run handler was not called")
	}
	if flagHost != "10.0.0.1" {
		t.Errorf("flagHost = %q, want %q", flagHost, "10.0.0.1")
	}
}

func TestParseAndRun_PositionalArgsPassedToRun(t *testing.T) {
	resetGlobalFlags(t)
	var capturedArgs []string
	cmd := &command{
		name: "test",
		use:  "test <file>",
		run: func(args []string) error {
			capturedArgs = args
			return nil
		},
	}
	err := parseAndRun(cmd, []string{"--name", "dev", "myfile.sql"}, "dabazo test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(capturedArgs) != 1 || capturedArgs[0] != "myfile.sql" {
		t.Errorf("capturedArgs = %v, want [myfile.sql]", capturedArgs)
	}
}

// --------------------------------------------------------------------------
// resolveAndRun — dispatch routing
// --------------------------------------------------------------------------

func TestResolveAndRun_DispatchesToLeafCommand(t *testing.T) {
	resetGlobalFlags(t)
	var ran bool
	leaf := &command{
		name: "leaf",
		use:  "leaf",
		run: func(args []string) error {
			ran = true
			return nil
		},
	}
	root := &command{
		name:        "root",
		subcommands: []*command{leaf},
	}
	if err := resolveAndRun(root, []string{"leaf"}, "root"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ran {
		t.Error("leaf run handler was not called")
	}
}

func TestResolveAndRun_TwoLevelDispatch(t *testing.T) {
	resetGlobalFlags(t)
	var ran bool
	leaf := &command{
		name: "user",
		use:  "user <username>",
		run: func(args []string) error {
			ran = true
			return nil
		},
	}
	group := &command{
		name:        "create",
		subcommands: []*command{leaf},
	}
	root := &command{
		name:        "root",
		subcommands: []*command{group},
	}
	if err := resolveAndRun(root, []string{"create", "user", "alice"}, "root"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ran {
		t.Error("leaf run handler was not called for two-level dispatch")
	}
}

func TestResolveAndRun_GroupWithNoArgsShowsHelp(t *testing.T) {
	resetGlobalFlags(t)
	leaf := &command{name: "sub", use: "sub", short: "A subcommand", run: func(args []string) error { return nil }}
	group := &command{
		name:        "create",
		short:       "Configuration",
		subcommands: []*command{leaf},
	}
	root := &command{
		name:        "root",
		subcommands: []*command{group},
	}
	// Passing zero args beyond "create" should trigger help output (not an error).
	err := resolveAndRun(root, []string{"create"}, "root")
	if err != nil {
		t.Errorf("expected nil error for group with no args, got: %v", err)
	}
}

func TestResolveAndRun_FlagBeforeSubcommandDoesNotDispatch(t *testing.T) {
	// If the first arg starts with '-', it should NOT be treated as a subcommand.
	resetGlobalFlags(t)
	var ranLeaf bool
	leaf := &command{
		name: "sub",
		use:  "sub",
		run:  func(args []string) error { ranLeaf = true; return nil },
	}
	var ranRoot bool
	root := &command{
		name: "root",
		use:  "root",
		run:  func(args []string) error { ranRoot = true; return nil },
		subcommands: []*command{leaf},
	}
	// "--name foo" starts with '-', so the dispatcher should NOT look for a subcommand;
	// since root has a run handler it should run root.
	if err := resolveAndRun(root, []string{"--name", "foo"}, "root"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ranLeaf {
		t.Error("leaf should NOT run when first arg is a flag")
	}
	if !ranRoot {
		t.Error("root run handler should be called when first arg is a flag")
	}
}

// --------------------------------------------------------------------------
// Full root-command dispatch via dispatch()
// --------------------------------------------------------------------------

func TestDispatch_InstallRouting(t *testing.T) {
	resetGlobalFlags(t)
	// dispatch() with "install --help" should not error (help exits normally).
	root := newRootCommand()
	// We can't easily intercept os.Exit in unit tests, so we just verify the
	// dispatch reaches the install command by checking help output does not panic.
	// parseAndRun detects --help and returns nil without calling run.
	err := dispatch(root, []string{"dabazo", "install", "--help"})
	if err != nil {
		t.Errorf("dispatch install --help returned error: %v", err)
	}
}

func TestDispatch_ConfigUserHelpRouting(t *testing.T) {
	resetGlobalFlags(t)
	root := newRootCommand()
	err := dispatch(root, []string{"dabazo", "create", "user", "--help"})
	if err != nil {
		t.Errorf("dispatch config user --help returned error: %v", err)
	}
}

func TestDispatch_RegistryAddHelpRouting(t *testing.T) {
	resetGlobalFlags(t)
	root := newRootCommand()
	err := dispatch(root, []string{"dabazo", "registry", "add", "--help"})
	if err != nil {
		t.Errorf("dispatch registry add --help returned error: %v", err)
	}
}

func TestDispatch_RegistryRemoveHelpRouting(t *testing.T) {
	resetGlobalFlags(t)
	root := newRootCommand()
	err := dispatch(root, []string{"dabazo", "registry", "remove", "--help"})
	if err != nil {
		t.Errorf("dispatch registry remove --help returned error: %v", err)
	}
}

func TestDispatch_RootHelpFlag(t *testing.T) {
	resetGlobalFlags(t)
	root := newRootCommand()
	// root has no run handler and no group command logic; resolveAndRun with --help
	// from root should hit handleGroupHelp which calls printHelp and returns nil.
	err := dispatch(root, []string{"dabazo", "--help"})
	if err != nil {
		t.Errorf("dispatch --help returned error: %v", err)
	}
}

// --------------------------------------------------------------------------
// newRootCommand structure
// --------------------------------------------------------------------------

func TestNewRootCommand_HasNineSubcommands(t *testing.T) {
	root := newRootCommand()
	want := 9
	got := len(root.subcommands)
	if got != want {
		t.Errorf("root has %d subcommands, want %d", got, want)
	}
}

func TestNewRootCommand_AllExpectedSubcommands(t *testing.T) {
	root := newRootCommand()
	names := make(map[string]bool)
	for _, sub := range root.subcommands {
		names[sub.name] = true
	}
	expected := []string{"install", "start", "stop", "create", "migrate", "snapshot", "registry", "list", "uninstall"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing subcommand %q in root command", name)
		}
	}
}

// --------------------------------------------------------------------------
// Command descriptor field completeness
// --------------------------------------------------------------------------

func TestCommandDescriptors_LeafCommandsHaveRunHandlers(t *testing.T) {
	root := newRootCommand()
	for _, sub := range root.subcommands {
		if len(sub.subcommands) > 0 {
			// group command — run should be nil
			if sub.run != nil {
				t.Errorf("group command %q should have nil run handler", sub.name)
			}
			for _, leaf := range sub.subcommands {
				if leaf.run == nil {
					t.Errorf("leaf command %q under %q has nil run handler", leaf.name, sub.name)
				}
			}
		} else {
			// leaf command
			if sub.run == nil {
				t.Errorf("leaf command %q has nil run handler", sub.name)
			}
		}
	}
}

func TestCommandDescriptors_AllCommandsHaveShortDescription(t *testing.T) {
	root := newRootCommand()
	var check func(cmd *command)
	check = func(cmd *command) {
		if cmd.short == "" && cmd.long == "" {
			t.Errorf("command %q has no short or long description", cmd.name)
		}
		for _, sub := range cmd.subcommands {
			check(sub)
		}
	}
	check(root)
}

func TestCommandDescriptors_LocalFlagsOnExpectedCommands(t *testing.T) {
	root := newRootCommand()

	migrateCmd := findSubcommand(root, "migrate")
	if migrateCmd == nil || migrateCmd.localFlags == nil {
		t.Error("migrate command should have localFlags (--user)")
	}

	snapshotCmd := findSubcommand(root, "snapshot")
	if snapshotCmd == nil || snapshotCmd.localFlags == nil {
		t.Error("snapshot command should have localFlags (--force)")
	}

	uninstallCmd := findSubcommand(root, "uninstall")
	if uninstallCmd == nil || uninstallCmd.localFlags == nil {
		t.Error("uninstall command should have localFlags (--purge)")
	}

	registryCmd := findSubcommand(root, "registry")
	if registryCmd == nil {
		t.Fatal("registry group command not found")
	}
	registryAddCmd := findSubcommand(registryCmd, "add")
	if registryAddCmd == nil || registryAddCmd.localFlags == nil {
		t.Error("registry add command should have localFlags (--host)")
	}
}

func TestCommandDescriptors_LocalFlagsAbsentOnSimpleCommands(t *testing.T) {
	// Commands without local flags should have nil localFlags.
	root := newRootCommand()
	for _, name := range []string{"install", "start", "stop", "list"} {
		cmd := findSubcommand(root, name)
		if cmd == nil {
			t.Errorf("command %q not found", name)
			continue
		}
		if cmd.localFlags != nil {
			t.Errorf("command %q should NOT have localFlags", name)
		}
	}
}

// --------------------------------------------------------------------------
// localFlags — flag registration verification
// --------------------------------------------------------------------------

func TestLocalFlags_MigrateRegistersUserFlag(t *testing.T) {
	resetGlobalFlags(t)
	cmd := newMigrateCommand()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cmd.localFlags(fs)
	if fs.Lookup("user") == nil {
		t.Error("migrate localFlags did not register 'user' flag")
	}
}

func TestLocalFlags_SnapshotRegistersForceFlag(t *testing.T) {
	resetGlobalFlags(t)
	cmd := newSnapshotCommand()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cmd.localFlags(fs)
	if fs.Lookup("force") == nil {
		t.Error("snapshot localFlags did not register 'force' flag")
	}
}

func TestLocalFlags_UninstallRegistersPurgeFlag(t *testing.T) {
	resetGlobalFlags(t)
	cmd := newUninstallCommand()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cmd.localFlags(fs)
	if fs.Lookup("purge") == nil {
		t.Error("uninstall localFlags did not register 'purge' flag")
	}
}

func TestLocalFlags_RegistryAddRegistersHostFlag(t *testing.T) {
	resetGlobalFlags(t)
	cmd := newRegistryAddCommand()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cmd.localFlags(fs)
	f := fs.Lookup("host")
	if f == nil {
		t.Fatal("registry add localFlags did not register 'host' flag")
	}
	if f.DefValue != "localhost" {
		t.Errorf("host flag default = %q, want %q", f.DefValue, "localhost")
	}
}

// --------------------------------------------------------------------------
// Flag parsing — combined global + local
// --------------------------------------------------------------------------

func TestFlagParsing_GlobalAndLocalCombined(t *testing.T) {
	resetGlobalFlags(t)
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	registerGlobalFlags(fs)
	fs.StringVar(&flagHost, "host", "localhost", "host address")

	args := []string{"--name", "test", "--engine", "postgres:16", "--port", "5432", "--host", "10.0.0.5"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if flagName != "test" {
		t.Errorf("flagName = %q, want %q", flagName, "test")
	}
	if flagEngine != "postgres:16" {
		t.Errorf("flagEngine = %q, want %q", flagEngine, "postgres:16")
	}
	if flagPort != 5432 {
		t.Errorf("flagPort = %d, want 5432", flagPort)
	}
	if flagHost != "10.0.0.5" {
		t.Errorf("flagHost = %q, want %q", flagHost, "10.0.0.5")
	}
}

func TestFlagParsing_YesShortAliasSetsFlag(t *testing.T) {
	resetGlobalFlags(t)
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	registerGlobalFlags(fs)
	if err := fs.Parse([]string{"-y"}); err != nil {
		t.Fatalf("unexpected parse error for -y: %v", err)
	}
	if !flagYes {
		t.Error("flagYes not set after -y")
	}
}

func TestFlagParsing_PortParsedAsInt(t *testing.T) {
	resetGlobalFlags(t)
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	registerGlobalFlags(fs)
	if err := fs.Parse([]string{"--port", "5433"}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if flagPort != 5433 {
		t.Errorf("flagPort = %d, want 5433", flagPort)
	}
}
