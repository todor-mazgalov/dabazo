// Package winget implements the PackageManager interface for Windows via winget.
package winget

// Manager implements PackageManager for winget.
type Manager struct{}

// Name returns "winget".
func (m *Manager) Name() string { return "winget" }

// InstallCommand returns the argv for installing packages via winget.
func (m *Manager) InstallCommand(pkgs []string) []string {
	args := []string{"winget", "install", "--accept-package-agreements", "--accept-source-agreements"}
	for _, pkg := range pkgs {
		args = append(args, "--id", pkg)
	}
	return args
}

// UninstallCommand returns the argv for removing packages via winget.
func (m *Manager) UninstallCommand(pkgs []string) []string {
	args := []string{"winget", "uninstall"}
	for _, pkg := range pkgs {
		args = append(args, "--id", pkg)
	}
	return args
}

// ServiceStart returns the argv for starting PostgreSQL via pg_ctl (Windows).
func (m *Manager) ServiceStart(svc string) []string {
	return []string{"pg_ctl", "start", "-D", svc}
}

// ServiceStop returns the argv for stopping PostgreSQL via pg_ctl (Windows).
func (m *Manager) ServiceStop(svc string) []string {
	return []string{"pg_ctl", "stop", "-D", svc}
}
