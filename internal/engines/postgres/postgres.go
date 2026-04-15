// Package postgres implements the Engine interface for PostgreSQL.
package postgres

import (
	"fmt"
	"net"
	"runtime"
	"strconv"
	"time"

	"github.com/todor-mazgalov/dabazo/internal/engines"
)

// Driver implements the Engine interface for PostgreSQL.
type Driver struct{}

// Name returns "postgres".
func (d *Driver) Name() string { return "postgres" }

// Plan returns an install plan for PostgreSQL on the detected platform.
func (d *Driver) Plan(version string, port int, pm engines.PackageManager) (engines.InstallPlan, error) {
	plan := engines.InstallPlan{
		EngineName: "postgres",
		Version:    version,
		Port:       port,
		PkgManager: pm.Name(),
	}

	switch pm.Name() {
	case "apt":
		planApt(&plan, version)
	case "dnf":
		planDnf(&plan, version)
	case "brew":
		planBrew(&plan, version)
	case "winget":
		planWinget(&plan, version)
	case "choco":
		planChoco(&plan, version)
	default:
		return plan, fmt.Errorf("unsupported package manager: %s", pm.Name())
	}

	return plan, nil
}

func planApt(plan *engines.InstallPlan, version string) {
	pkgs := []string{
		fmt.Sprintf("postgresql-%s", version),
		fmt.Sprintf("postgresql-client-%s", version),
	}
	plan.Packages = pkgs
	plan.Source = "deb.debian.org (main)"
	plan.Commands = [][]string{
		{"sudo", "apt-get", "update"},
		{"sudo", "apt-get", "install", "-y", pkgs[0], pkgs[1]},
	}
	plan.DataDir = fmt.Sprintf("/var/lib/postgresql/%s/main", version)
	plan.ServiceName = fmt.Sprintf("postgresql@%s-main", version)
	plan.PostInstall = []string{
		fmt.Sprintf("initdb data directory: /var/lib/postgresql/%s/main", version),
	}
}

func planDnf(plan *engines.InstallPlan, version string) {
	pkgs := []string{
		fmt.Sprintf("postgresql%s-server", version),
		fmt.Sprintf("postgresql%s", version),
	}
	plan.Packages = pkgs
	plan.Source = "fedora/rhel repositories"
	plan.Commands = [][]string{
		{"sudo", "dnf", "install", "-y", pkgs[0], pkgs[1]},
		{"sudo", fmt.Sprintf("/usr/pgsql-%s/bin/postgresql-%s-setup", version, version), "initdb"},
	}
	plan.DataDir = fmt.Sprintf("/var/lib/pgsql/%s/data", version)
	plan.ServiceName = fmt.Sprintf("postgresql-%s", version)
	plan.PostInstall = []string{
		fmt.Sprintf("initdb data directory: /var/lib/pgsql/%s/data", version),
	}
}

func planBrew(plan *engines.InstallPlan, version string) {
	pkg := fmt.Sprintf("postgresql@%s", version)
	plan.Packages = []string{pkg}
	plan.Source = "homebrew/core"
	plan.Commands = [][]string{
		{"brew", "install", pkg},
	}
	plan.DataDir = fmt.Sprintf("/usr/local/var/postgresql@%s", version)
	plan.ServiceName = pkg
	plan.BinDir = fmt.Sprintf("/usr/local/opt/postgresql@%s/bin", version)
	plan.PostInstall = []string{
		fmt.Sprintf("initdb data directory: /usr/local/var/postgresql@%s", version),
	}
}

func planWinget(plan *engines.InstallPlan, version string) {
	plan.Packages = []string{fmt.Sprintf("PostgreSQL.PostgreSQL.%s", version)}
	plan.Source = "winget"
	plan.Commands = [][]string{
		{"winget", "install", "--accept-package-agreements", "--accept-source-agreements",
			"--id", fmt.Sprintf("PostgreSQL.PostgreSQL.%s", version)},
	}
	plan.DataDir = fmt.Sprintf("C:\\Program Files\\PostgreSQL\\%s\\data", version)
	plan.BinDir = fmt.Sprintf("C:\\Program Files\\PostgreSQL\\%s\\bin", version)
	plan.PostInstall = []string{
		fmt.Sprintf("initdb data directory: C:\\Program Files\\PostgreSQL\\%s\\data", version),
	}
}

func planChoco(plan *engines.InstallPlan, version string) {
	plan.Packages = []string{"postgresql"}
	plan.Source = "chocolatey"
	plan.Commands = [][]string{
		{"choco", "install", "-y", "postgresql", "--version", version},
	}
	plan.DataDir = fmt.Sprintf("C:\\Program Files\\PostgreSQL\\%s\\data", version)
	plan.BinDir = fmt.Sprintf("C:\\Program Files\\PostgreSQL\\%s\\bin", version)
	plan.PostInstall = []string{
		fmt.Sprintf("initdb data directory: C:\\Program Files\\PostgreSQL\\%s\\data", version),
	}
}

// Install executes the install plan by running each command in sequence.
func (d *Driver) Install(plan engines.InstallPlan, runner engines.CommandRunner) error {
	for _, cmd := range plan.Commands {
		if len(cmd) == 0 {
			continue
		}
		if _, err := runner.Run(cmd[0], cmd[1:]...); err != nil {
			return fmt.Errorf("install step %q: %w", cmd[0], err)
		}
	}
	if err := d.initDBIfNeeded(plan, runner); err != nil {
		return fmt.Errorf("post-install init: %w", err)
	}
	if err := d.configurePort(plan, runner); err != nil {
		return fmt.Errorf("configuring port: %w", err)
	}
	return nil
}

// initDBIfNeeded runs initdb if the data directory does not yet exist.
func (d *Driver) initDBIfNeeded(plan engines.InstallPlan, runner engines.CommandRunner) error {
	if plan.PkgManager == "apt" {
		return nil
	}
	initdb := "initdb"
	if plan.BinDir != "" {
		initdb = plan.BinDir + "/initdb"
		if runtime.GOOS == "windows" {
			initdb = plan.BinDir + "\\initdb.exe"
		}
	}
	_, err := runner.Run(initdb, "-D", plan.DataDir, "-U", "postgres", "-E", "UTF8")
	if err != nil {
		return fmt.Errorf("initdb: %w", err)
	}
	return nil
}

// configurePort sets the listen port in postgresql.conf.
func (d *Driver) configurePort(plan engines.InstallPlan, runner engines.CommandRunner) error {
	confPath := plan.DataDir + "/postgresql.conf"
	portLine := fmt.Sprintf("port = %d", plan.Port)
	if runtime.GOOS == "windows" {
		_, err := runner.Run("powershell", "-Command",
			fmt.Sprintf(`(Get-Content '%s') -replace '^#?port\s*=.*', '%s' | Set-Content '%s'`,
				confPath, portLine, confPath))
		return err
	}
	_, err := runner.Run("sudo", "sed", "-i",
		fmt.Sprintf(`s/^#\?port\s*=.*/port = %d/`, plan.Port),
		confPath)
	return err
}

// Start starts the PostgreSQL service for the given instance.
func (d *Driver) Start(inst engines.Instance, runner engines.CommandRunner) error {
	if d.IsRunning(inst) {
		fmt.Println("already running")
		return nil
	}
	cmd := d.startCommand(inst)
	if _, err := runner.Run(cmd[0], cmd[1:]...); err != nil {
		return fmt.Errorf("starting postgres: %w", err)
	}
	return nil
}

// Stop stops the PostgreSQL service for the given instance.
func (d *Driver) Stop(inst engines.Instance, runner engines.CommandRunner) error {
	if !d.IsRunning(inst) {
		fmt.Println("already stopped")
		return nil
	}
	cmd := d.stopCommand(inst)
	if _, err := runner.Run(cmd[0], cmd[1:]...); err != nil {
		return fmt.Errorf("stopping postgres: %w", err)
	}
	return nil
}

func (d *Driver) startCommand(inst engines.Instance) []string {
	if runtime.GOOS == "windows" || inst.BinDir != "" {
		pgctl := pgCtlPath(inst)
		return []string{pgctl, "start", "-D", inst.DataDir, "-w"}
	}
	switch inst.PackageManager {
	case "brew":
		return []string{"brew", "services", "start", inst.ServiceName}
	default:
		return []string{"sudo", "systemctl", "start", inst.ServiceName}
	}
}

func (d *Driver) stopCommand(inst engines.Instance) []string {
	if runtime.GOOS == "windows" || inst.BinDir != "" {
		pgctl := pgCtlPath(inst)
		return []string{pgctl, "stop", "-D", inst.DataDir, "-w"}
	}
	switch inst.PackageManager {
	case "brew":
		return []string{"brew", "services", "stop", inst.ServiceName}
	default:
		return []string{"sudo", "systemctl", "stop", inst.ServiceName}
	}
}

func pgCtlPath(inst engines.Instance) string {
	if inst.BinDir != "" {
		sep := "/"
		if runtime.GOOS == "windows" {
			sep = "\\"
		}
		return inst.BinDir + sep + "pg_ctl"
	}
	return "pg_ctl"
}

// CreateUser creates a PostgreSQL role and database using psql.
func (d *Driver) CreateUser(inst engines.Instance, username, password string, runner engines.CommandRunner) error {
	psql := psqlPath(inst)
	sql := fmt.Sprintf(
		"CREATE ROLE %s WITH LOGIN PASSWORD '%s'; CREATE DATABASE %s OWNER %s;",
		username, password, username, username,
	)
	_, err := runner.Run(psql, "-h", inst.Host, "-p", strconv.Itoa(inst.Port),
		"-U", "postgres", "-c", sql)
	if err != nil {
		return fmt.Errorf("creating user %q: %w", username, err)
	}
	return nil
}

// ApplySQL executes a SQL file against the instance database.
func (d *Driver) ApplySQL(inst engines.Instance, user, password, filepath string, runner engines.CommandRunner) error {
	psql := psqlPath(inst)
	env := []string{fmt.Sprintf("PGPASSWORD=%s", password)}
	_, err := runner.RunWithEnv(env, psql, "-h", inst.Host, "-p", strconv.Itoa(inst.Port),
		"-U", user, "-d", user, "-f", filepath)
	if err != nil {
		return fmt.Errorf("applying SQL %q: %w", filepath, err)
	}
	return nil
}

// Dump exports a database to a plain SQL file using pg_dump.
func (d *Driver) Dump(inst engines.Instance, db, user, password, outPath string, runner engines.CommandRunner) error {
	pgdump := "pg_dump"
	if inst.BinDir != "" {
		sep := "/"
		if runtime.GOOS == "windows" {
			sep = "\\"
		}
		pgdump = inst.BinDir + sep + "pg_dump"
	}
	env := []string{fmt.Sprintf("PGPASSWORD=%s", password)}
	_, err := runner.RunWithEnv(env, pgdump,
		"-h", inst.Host, "-p", strconv.Itoa(inst.Port),
		"-U", user, "-Fp", "-f", outPath, db)
	if err != nil {
		return fmt.Errorf("dumping database %q: %w", db, err)
	}
	return nil
}

// IsRunning checks if PostgreSQL is accepting connections on the instance port.
func (d *Driver) IsRunning(inst engines.Instance) bool {
	addr := net.JoinHostPort(inst.Host, strconv.Itoa(inst.Port))
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// UninstallPlan returns the commands needed to uninstall PostgreSQL packages.
func (d *Driver) UninstallPlan(inst engines.Instance, pm engines.PackageManager) (engines.InstallPlan, error) {
	plan := engines.InstallPlan{
		EngineName: "postgres",
		Version:    inst.Version,
		Port:       inst.Port,
		PkgManager: pm.Name(),
	}

	switch pm.Name() {
	case "apt":
		pkgs := []string{
			fmt.Sprintf("postgresql-%s", inst.Version),
			fmt.Sprintf("postgresql-client-%s", inst.Version),
		}
		plan.Packages = pkgs
		plan.Commands = [][]string{
			{"sudo", "apt-get", "remove", "-y", pkgs[0], pkgs[1]},
		}
	case "dnf":
		pkgs := []string{
			fmt.Sprintf("postgresql%s-server", inst.Version),
			fmt.Sprintf("postgresql%s", inst.Version),
		}
		plan.Packages = pkgs
		plan.Commands = [][]string{
			{"sudo", "dnf", "remove", "-y", pkgs[0], pkgs[1]},
		}
	case "brew":
		pkg := fmt.Sprintf("postgresql@%s", inst.Version)
		plan.Packages = []string{pkg}
		plan.Commands = [][]string{
			{"brew", "uninstall", pkg},
		}
	case "winget":
		pkg := fmt.Sprintf("PostgreSQL.PostgreSQL.%s", inst.Version)
		plan.Packages = []string{pkg}
		plan.Commands = [][]string{
			{"winget", "uninstall", "--id", pkg},
		}
	case "choco":
		plan.Packages = []string{"postgresql"}
		plan.Commands = [][]string{
			{"choco", "uninstall", "-y", "postgresql"},
		}
	default:
		return plan, fmt.Errorf("unsupported package manager: %s", pm.Name())
	}

	return plan, nil
}

func psqlPath(inst engines.Instance) string {
	if inst.BinDir != "" {
		sep := "/"
		if runtime.GOOS == "windows" {
			sep = "\\"
		}
		return inst.BinDir + sep + "psql"
	}
	return "psql"
}
