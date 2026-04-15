package postgres

import (
	"fmt"
	"strings"
	"testing"

	"github.com/todor-mazgalov/dabazo/internal/engines"
)

// mockPM is a minimal PackageManager mock for testing Plan().
type mockPM struct {
	name string
}

func (m *mockPM) Name() string                          { return m.name }
func (m *mockPM) InstallCommand(pkgs []string) []string { return nil }
func (m *mockPM) UninstallCommand(pkgs []string) []string { return nil }
func (m *mockPM) ServiceStart(svc string) []string      { return nil }
func (m *mockPM) ServiceStop(svc string) []string       { return nil }

func TestDriver_Name(t *testing.T) {
	d := &Driver{}
	if got := d.Name(); got != "postgres" {
		t.Errorf("Name() = %q, want %q", got, "postgres")
	}
}

func TestDriver_Plan_Apt(t *testing.T) {
	d := &Driver{}
	pm := &mockPM{name: "apt"}
	plan, err := d.Plan("16", 5432, pm)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if plan.EngineName != "postgres" {
		t.Errorf("EngineName = %q, want %q", plan.EngineName, "postgres")
	}
	if plan.Version != "16" {
		t.Errorf("Version = %q, want %q", plan.Version, "16")
	}
	if plan.Port != 5432 {
		t.Errorf("Port = %d, want %d", plan.Port, 5432)
	}
	if plan.PkgManager != "apt" {
		t.Errorf("PkgManager = %q, want %q", plan.PkgManager, "apt")
	}
	if plan.Source != "deb.debian.org (main)" {
		t.Errorf("Source = %q, want %q", plan.Source, "deb.debian.org (main)")
	}

	// Check packages
	wantPkgs := []string{"postgresql-16", "postgresql-client-16"}
	if len(plan.Packages) != len(wantPkgs) {
		t.Fatalf("Packages count = %d, want %d", len(plan.Packages), len(wantPkgs))
	}
	for i, pkg := range wantPkgs {
		if plan.Packages[i] != pkg {
			t.Errorf("Packages[%d] = %q, want %q", i, plan.Packages[i], pkg)
		}
	}

	// Check commands include apt-get update and install
	if len(plan.Commands) != 2 {
		t.Fatalf("Commands count = %d, want 2", len(plan.Commands))
	}
	if plan.Commands[0][1] != "apt-get" || plan.Commands[0][2] != "update" {
		t.Errorf("first command should be apt-get update, got %v", plan.Commands[0])
	}
	if plan.Commands[1][1] != "apt-get" || plan.Commands[1][2] != "install" {
		t.Errorf("second command should be apt-get install, got %v", plan.Commands[1])
	}

	// Check data dir and service name
	if plan.DataDir != "/var/lib/postgresql/16/main" {
		t.Errorf("DataDir = %q, want %q", plan.DataDir, "/var/lib/postgresql/16/main")
	}
	if plan.ServiceName != "postgresql@16-main" {
		t.Errorf("ServiceName = %q, want %q", plan.ServiceName, "postgresql@16-main")
	}
}

func TestDriver_Plan_Dnf(t *testing.T) {
	d := &Driver{}
	pm := &mockPM{name: "dnf"}
	plan, err := d.Plan("16", 5433, pm)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	wantPkgs := []string{"postgresql16-server", "postgresql16"}
	if len(plan.Packages) != len(wantPkgs) {
		t.Fatalf("Packages count = %d, want %d", len(plan.Packages), len(wantPkgs))
	}
	for i, pkg := range wantPkgs {
		if plan.Packages[i] != pkg {
			t.Errorf("Packages[%d] = %q, want %q", i, plan.Packages[i], pkg)
		}
	}
	if plan.Source != "fedora/rhel repositories" {
		t.Errorf("Source = %q, want %q", plan.Source, "fedora/rhel repositories")
	}
	if plan.DataDir != "/var/lib/pgsql/16/data" {
		t.Errorf("DataDir = %q, want %q", plan.DataDir, "/var/lib/pgsql/16/data")
	}
	if plan.ServiceName != "postgresql-16" {
		t.Errorf("ServiceName = %q, want %q", plan.ServiceName, "postgresql-16")
	}
}

func TestDriver_Plan_Brew(t *testing.T) {
	d := &Driver{}
	pm := &mockPM{name: "brew"}
	plan, err := d.Plan("16", 5432, pm)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(plan.Packages) != 1 || plan.Packages[0] != "postgresql@16" {
		t.Errorf("Packages = %v, want [postgresql@16]", plan.Packages)
	}
	if plan.Source != "homebrew/core" {
		t.Errorf("Source = %q, want %q", plan.Source, "homebrew/core")
	}
	if plan.ServiceName != "postgresql@16" {
		t.Errorf("ServiceName = %q, want %q", plan.ServiceName, "postgresql@16")
	}
	if plan.BinDir == "" {
		t.Error("BinDir should be set for brew")
	}
}

func TestDriver_Plan_Winget(t *testing.T) {
	d := &Driver{}
	pm := &mockPM{name: "winget"}
	plan, err := d.Plan("16", 5432, pm)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(plan.Packages) != 1 || plan.Packages[0] != "PostgreSQL.PostgreSQL.16" {
		t.Errorf("Packages = %v, want [PostgreSQL.PostgreSQL.16]", plan.Packages)
	}
	if plan.Source != "winget" {
		t.Errorf("Source = %q, want %q", plan.Source, "winget")
	}
	if plan.BinDir == "" {
		t.Error("BinDir should be set for winget")
	}
	if plan.DataDir == "" {
		t.Error("DataDir should be set for winget")
	}
}

func TestDriver_Plan_Choco(t *testing.T) {
	d := &Driver{}
	pm := &mockPM{name: "choco"}
	plan, err := d.Plan("16", 5432, pm)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(plan.Packages) != 1 || plan.Packages[0] != "postgresql" {
		t.Errorf("Packages = %v, want [postgresql]", plan.Packages)
	}
	if plan.Source != "chocolatey" {
		t.Errorf("Source = %q, want %q", plan.Source, "chocolatey")
	}
}

func TestDriver_Plan_UnsupportedPM(t *testing.T) {
	d := &Driver{}
	pm := &mockPM{name: "unknown"}
	_, err := d.Plan("16", 5432, pm)
	if err == nil {
		t.Error("expected error for unsupported package manager")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("error should mention unsupported, got: %v", err)
	}
}

func TestDriver_Plan_VersionPropagation(t *testing.T) {
	d := &Driver{}
	versions := []string{"14", "15", "16", "17"}
	for _, v := range versions {
		t.Run("version_"+v, func(t *testing.T) {
			pm := &mockPM{name: "apt"}
			plan, err := d.Plan(v, 5432, pm)
			if err != nil {
				t.Fatalf("Plan: %v", err)
			}
			if plan.Version != v {
				t.Errorf("Version = %q, want %q", plan.Version, v)
			}
			// Package names should contain the version
			for _, pkg := range plan.Packages {
				if !strings.Contains(pkg, v) {
					t.Errorf("package %q should contain version %q", pkg, v)
				}
			}
		})
	}
}

func TestDriver_Plan_PortPropagation(t *testing.T) {
	d := &Driver{}
	pm := &mockPM{name: "apt"}
	plan, err := d.Plan("16", 9999, pm)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if plan.Port != 9999 {
		t.Errorf("Port = %d, want %d", plan.Port, 9999)
	}
}

// mockRunner records commands for testing Install, CreateUser, ApplySQL, Dump.
type mockRunner struct {
	calls   []string
	failOn  string
	outputs map[string][]byte
}

func (r *mockRunner) Run(name string, args ...string) ([]byte, error) {
	cmd := name + " " + strings.Join(args, " ")
	r.calls = append(r.calls, cmd)
	if r.failOn != "" && strings.Contains(cmd, r.failOn) {
		return nil, fmt.Errorf("mock failure on %q", r.failOn)
	}
	if r.outputs != nil {
		if out, ok := r.outputs[name]; ok {
			return out, nil
		}
	}
	return nil, nil
}

func (r *mockRunner) RunWithEnv(env []string, name string, args ...string) ([]byte, error) {
	cmd := name + " " + strings.Join(args, " ")
	r.calls = append(r.calls, "ENV:"+strings.Join(env, ",")+"|"+cmd)
	if r.failOn != "" && strings.Contains(cmd, r.failOn) {
		return nil, fmt.Errorf("mock failure on %q", r.failOn)
	}
	return nil, nil
}

func TestDriver_Install_RunsAllCommands(t *testing.T) {
	d := &Driver{}
	plan := engines.InstallPlan{
		PkgManager: "apt", // apt skips initdb
		Commands: [][]string{
			{"sudo", "apt-get", "update"},
			{"sudo", "apt-get", "install", "-y", "postgresql-16"},
		},
		DataDir: "/var/lib/postgresql/16/main",
		Port:    5432,
	}
	runner := &mockRunner{}
	err := d.Install(plan, runner)
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	// Should have run both commands plus configurePort (sed)
	if len(runner.calls) < 2 {
		t.Errorf("expected at least 2 calls, got %d: %v", len(runner.calls), runner.calls)
	}
}

func TestDriver_Install_FailsOnCommandError(t *testing.T) {
	d := &Driver{}
	plan := engines.InstallPlan{
		PkgManager: "apt",
		Commands: [][]string{
			{"sudo", "apt-get", "update"},
		},
		DataDir: "/var/lib/postgresql/16/main",
		Port:    5432,
	}
	runner := &mockRunner{failOn: "apt-get"}
	err := d.Install(plan, runner)
	if err == nil {
		t.Error("expected error when command fails")
	}
}

func TestDriver_CreateUser_CallsPsql(t *testing.T) {
	d := &Driver{}
	inst := engines.Instance{
		Host: "localhost",
		Port: 5432,
	}
	runner := &mockRunner{}
	err := d.CreateUser(inst, "alice", "secret123", runner)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if len(runner.calls) == 0 {
		t.Fatal("expected psql call")
	}
	call := runner.calls[0]
	if !strings.Contains(call, "psql") {
		t.Errorf("expected psql in command, got: %s", call)
	}
	if !strings.Contains(call, "alice") {
		t.Errorf("expected username in command, got: %s", call)
	}
	if !strings.Contains(call, "secret123") {
		t.Errorf("expected password in SQL, got: %s", call)
	}
}

func TestDriver_CreateUser_Error(t *testing.T) {
	d := &Driver{}
	inst := engines.Instance{Host: "localhost", Port: 5432}
	runner := &mockRunner{failOn: "psql"}
	err := d.CreateUser(inst, "alice", "pass", runner)
	if err == nil {
		t.Error("expected error when psql fails")
	}
}

func TestDriver_ApplySQL_UsesPGPASSWORD(t *testing.T) {
	d := &Driver{}
	inst := engines.Instance{Host: "localhost", Port: 5432}
	runner := &mockRunner{}
	err := d.ApplySQL(inst, "alice", "pass123", "", "", "/tmp/test.sql", runner)
	if err != nil {
		t.Fatalf("ApplySQL: %v", err)
	}
	if len(runner.calls) == 0 {
		t.Fatal("expected psql call")
	}
	call := runner.calls[0]
	if !strings.Contains(call, "PGPASSWORD=pass123") {
		t.Errorf("expected PGPASSWORD in env, got: %s", call)
	}
	if !strings.Contains(call, "/tmp/test.sql") {
		t.Errorf("expected filepath in command, got: %s", call)
	}
}

func TestDriver_Dump_UsesPGPASSWORD(t *testing.T) {
	d := &Driver{}
	inst := engines.Instance{Host: "localhost", Port: 5432}
	runner := &mockRunner{}
	err := d.Dump(inst, "mydb", "alice", "pass123", "/tmp/dump.sql", runner)
	if err != nil {
		t.Fatalf("Dump: %v", err)
	}
	if len(runner.calls) == 0 {
		t.Fatal("expected pg_dump call")
	}
	call := runner.calls[0]
	if !strings.Contains(call, "PGPASSWORD=pass123") {
		t.Errorf("expected PGPASSWORD in env, got: %s", call)
	}
	if !strings.Contains(call, "pg_dump") {
		t.Errorf("expected pg_dump in command, got: %s", call)
	}
	if !strings.Contains(call, "mydb") {
		t.Errorf("expected db name in command, got: %s", call)
	}
	if !strings.Contains(call, "-Fp") {
		t.Errorf("expected plain format flag -Fp, got: %s", call)
	}
}

func TestDriver_Dump_WithBinDir(t *testing.T) {
	d := &Driver{}
	inst := engines.Instance{Host: "localhost", Port: 5432, BinDir: "/usr/local/pgsql/bin"}
	runner := &mockRunner{}
	err := d.Dump(inst, "mydb", "u", "p", "/tmp/out.sql", runner)
	if err != nil {
		t.Fatalf("Dump: %v", err)
	}
	call := runner.calls[0]
	if !strings.Contains(call, "/usr/local/pgsql/bin") {
		t.Errorf("expected BinDir in pg_dump path, got: %s", call)
	}
}

func TestDriver_Dump_Error(t *testing.T) {
	d := &Driver{}
	inst := engines.Instance{Host: "localhost", Port: 5432}
	runner := &mockRunner{failOn: "pg_dump"}
	err := d.Dump(inst, "mydb", "u", "p", "/tmp/out.sql", runner)
	if err == nil {
		t.Error("expected error when pg_dump fails")
	}
}

func TestDriver_UninstallPlan_Apt(t *testing.T) {
	d := &Driver{}
	inst := engines.Instance{Version: "16", Port: 5432}
	pm := &mockPM{name: "apt"}
	plan, err := d.UninstallPlan(inst, pm)
	if err != nil {
		t.Fatalf("UninstallPlan: %v", err)
	}
	if len(plan.Packages) != 2 {
		t.Errorf("expected 2 packages, got %d", len(plan.Packages))
	}
	if len(plan.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(plan.Commands))
	}
	cmd := plan.Commands[0]
	if cmd[1] != "apt-get" || cmd[2] != "remove" {
		t.Errorf("expected apt-get remove, got %v", cmd)
	}
}

func TestDriver_UninstallPlan_Brew(t *testing.T) {
	d := &Driver{}
	inst := engines.Instance{Version: "16", Port: 5432}
	pm := &mockPM{name: "brew"}
	plan, err := d.UninstallPlan(inst, pm)
	if err != nil {
		t.Fatalf("UninstallPlan: %v", err)
	}
	if len(plan.Packages) != 1 || plan.Packages[0] != "postgresql@16" {
		t.Errorf("Packages = %v, want [postgresql@16]", plan.Packages)
	}
	cmd := plan.Commands[0]
	if cmd[0] != "brew" || cmd[1] != "uninstall" {
		t.Errorf("expected brew uninstall, got %v", cmd)
	}
}

func TestDriver_UninstallPlan_UnsupportedPM(t *testing.T) {
	d := &Driver{}
	inst := engines.Instance{Version: "16", Port: 5432}
	pm := &mockPM{name: "unknown"}
	_, err := d.UninstallPlan(inst, pm)
	if err == nil {
		t.Error("expected error for unsupported PM")
	}
}
