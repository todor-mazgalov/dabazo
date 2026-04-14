// Package engines defines the database engine driver interface and shared types.
package engines

// Instance represents a registered database instance.
type Instance struct {
	Name           string `json:"name"`
	Engine         string `json:"engine"`
	Version        string `json:"version"`
	Port           int    `json:"port"`
	Host           string `json:"host"`
	InstalledAt    string `json:"installedAt"`
	PackageManager string `json:"packageManager"`
	ServiceName    string `json:"serviceName,omitempty"`
	DataDir        string `json:"dataDir,omitempty"`
	BinDir         string `json:"binDir,omitempty"`
}

// InstallPlan describes the steps needed to install a database engine.
type InstallPlan struct {
	EngineName  string
	Version     string
	Port        int
	PkgManager  string
	Source      string
	Packages    []string
	Commands    [][]string
	PostInstall []string
	DataDir     string
	ServiceName string
	BinDir      string
}

// Engine is the interface that database engine drivers must implement.
type Engine interface {
	// Name returns the canonical engine name, e.g. "postgres".
	Name() string

	// Plan returns the install plan for the given version/port/OS without
	// executing anything. Used for the confirmation prompt.
	Plan(version string, port int, pm PackageManager) (InstallPlan, error)

	// Install executes a plan previously produced by Plan.
	Install(plan InstallPlan, runner CommandRunner) error

	// Start starts the database service for the given instance.
	Start(inst Instance, runner CommandRunner) error

	// Stop stops the database service for the given instance.
	Stop(inst Instance, runner CommandRunner) error

	// CreateUser creates a database role and database with the given credentials.
	CreateUser(inst Instance, username, password string, runner CommandRunner) error

	// ApplySQL applies a SQL file to the instance's database.
	ApplySQL(inst Instance, user, password, filepath string, runner CommandRunner) error

	// Dump exports a database to a plain SQL file.
	Dump(inst Instance, db, user, password, outPath string, runner CommandRunner) error

	// IsRunning checks whether the instance is currently accepting connections.
	IsRunning(inst Instance) bool

	// UninstallPlan returns the commands needed to uninstall the packages.
	UninstallPlan(inst Instance, pm PackageManager) (InstallPlan, error)
}

// PackageManager abstracts OS-specific package management.
type PackageManager interface {
	// Name returns the package manager identifier, e.g. "apt".
	Name() string

	// InstallCommand returns the full argv to install the given packages.
	InstallCommand(pkgs []string) []string

	// UninstallCommand returns the full argv to uninstall the given packages.
	UninstallCommand(pkgs []string) []string

	// ServiceStart returns the full argv to start a service by name.
	ServiceStart(svc string) []string

	// ServiceStop returns the full argv to stop a service by name.
	ServiceStop(svc string) []string
}

// CommandRunner abstracts command execution for testability.
type CommandRunner interface {
	// Run executes a command and returns combined output and any error.
	Run(name string, args ...string) ([]byte, error)

	// RunWithEnv executes a command with extra environment variables.
	RunWithEnv(env []string, name string, args ...string) ([]byte, error)
}
