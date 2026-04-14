// Package choco implements the PackageManager interface for Windows via Chocolatey.
package choco

// Manager implements PackageManager for choco.
type Manager struct{}

// Name returns "choco".
func (m *Manager) Name() string { return "choco" }

// InstallCommand returns the argv for installing packages via choco.
func (m *Manager) InstallCommand(pkgs []string) []string {
	args := []string{"choco", "install", "-y"}
	return append(args, pkgs...)
}

// UninstallCommand returns the argv for removing packages via choco.
func (m *Manager) UninstallCommand(pkgs []string) []string {
	args := []string{"choco", "uninstall", "-y"}
	return append(args, pkgs...)
}

// ServiceStart returns the argv for starting PostgreSQL via pg_ctl (Windows).
func (m *Manager) ServiceStart(svc string) []string {
	return []string{"pg_ctl", "start", "-D", svc}
}

// ServiceStop returns the argv for stopping PostgreSQL via pg_ctl (Windows).
func (m *Manager) ServiceStop(svc string) []string {
	return []string{"pg_ctl", "stop", "-D", svc}
}
