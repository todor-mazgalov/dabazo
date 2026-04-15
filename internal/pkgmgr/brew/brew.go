// Package brew implements the PackageManager interface for macOS via Homebrew.
package brew

// Manager implements PackageManager for brew.
type Manager struct{}

// Name returns "brew".
func (m *Manager) Name() string { return "brew" }

// InstallCommand returns the argv for installing packages via brew.
func (m *Manager) InstallCommand(pkgs []string) []string {
	args := []string{"brew", "install"}
	return append(args, pkgs...)
}

// UninstallCommand returns the argv for removing packages via brew.
func (m *Manager) UninstallCommand(pkgs []string) []string {
	args := []string{"brew", "uninstall"}
	return append(args, pkgs...)
}

// ServiceStart returns the argv for starting a brew service.
func (m *Manager) ServiceStart(svc string) []string {
	return []string{"brew", "services", "start", svc}
}

// ServiceStop returns the argv for stopping a brew service.
func (m *Manager) ServiceStop(svc string) []string {
	return []string{"brew", "services", "stop", svc}
}
