// Package pkgmgr provides OS detection and package manager resolution.
package pkgmgr

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/todor-mazgalov/dabazo/internal/engines"
	"github.com/todor-mazgalov/dabazo/internal/pkgmgr/apt"
	"github.com/todor-mazgalov/dabazo/internal/pkgmgr/brew"
	"github.com/todor-mazgalov/dabazo/internal/pkgmgr/choco"
	"github.com/todor-mazgalov/dabazo/internal/pkgmgr/dnf"
	"github.com/todor-mazgalov/dabazo/internal/pkgmgr/winget"
)

// Detect returns the appropriate package manager for the current OS.
func Detect() (engines.PackageManager, error) {
	switch runtime.GOOS {
	case "linux":
		return detectLinux()
	case "darwin":
		return &brew.Manager{}, nil
	case "windows":
		return detectWindows()
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// detectLinux checks for apt-get or dnf.
func detectLinux() (engines.PackageManager, error) {
	if _, err := exec.LookPath("apt-get"); err == nil {
		return &apt.Manager{}, nil
	}
	if _, err := exec.LookPath("dnf"); err == nil {
		return &dnf.Manager{}, nil
	}
	return nil, fmt.Errorf("no supported package manager found (need apt-get or dnf)")
}

// detectWindows checks for winget or choco.
func detectWindows() (engines.PackageManager, error) {
	if _, err := exec.LookPath("winget"); err == nil {
		return &winget.Manager{}, nil
	}
	if _, err := exec.LookPath("choco"); err == nil {
		return &choco.Manager{}, nil
	}
	return nil, fmt.Errorf("no supported package manager found (need winget or choco)")
}

// ByName returns a package manager by its registry name.
// Used when loading instances that record which PM was used at install time.
func ByName(name string) (engines.PackageManager, error) {
	switch name {
	case "apt":
		return &apt.Manager{}, nil
	case "dnf":
		return &dnf.Manager{}, nil
	case "brew":
		return &brew.Manager{}, nil
	case "winget":
		return &winget.Manager{}, nil
	case "choco":
		return &choco.Manager{}, nil
	case "external":
		return nil, fmt.Errorf("instance was added via `registry add`; dabazo does not manage its lifecycle")
	default:
		return nil, fmt.Errorf("unknown package manager: %s", name)
	}
}
