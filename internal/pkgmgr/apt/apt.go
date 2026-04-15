// Package apt implements the PackageManager interface for Debian/Ubuntu systems.
package apt

// Manager implements PackageManager for apt-get.
type Manager struct{}

// Name returns "apt".
func (m *Manager) Name() string { return "apt" }

// InstallCommand returns the argv for installing packages via apt-get.
func (m *Manager) InstallCommand(pkgs []string) []string {
	args := []string{"sudo", "apt-get", "install", "-y"}
	return append(args, pkgs...)
}

// UninstallCommand returns the argv for removing packages via apt-get.
func (m *Manager) UninstallCommand(pkgs []string) []string {
	args := []string{"sudo", "apt-get", "remove", "-y"}
	return append(args, pkgs...)
}

// ServiceStart returns the argv for starting a systemd service.
func (m *Manager) ServiceStart(svc string) []string {
	return []string{"sudo", "systemctl", "start", svc}
}

// ServiceStop returns the argv for stopping a systemd service.
func (m *Manager) ServiceStop(svc string) []string {
	return []string{"sudo", "systemctl", "stop", svc}
}
