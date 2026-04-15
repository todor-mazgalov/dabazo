// Package dnf implements the PackageManager interface for RHEL/Fedora systems.
package dnf

// Manager implements PackageManager for dnf.
type Manager struct{}

// Name returns "dnf".
func (m *Manager) Name() string { return "dnf" }

// InstallCommand returns the argv for installing packages via dnf.
func (m *Manager) InstallCommand(pkgs []string) []string {
	args := []string{"sudo", "dnf", "install", "-y"}
	return append(args, pkgs...)
}

// UninstallCommand returns the argv for removing packages via dnf.
func (m *Manager) UninstallCommand(pkgs []string) []string {
	args := []string{"sudo", "dnf", "remove", "-y"}
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
