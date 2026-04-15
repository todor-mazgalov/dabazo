package pkgmgr

import (
	"strings"
	"testing"
)

func TestByName_ValidManagers(t *testing.T) {
	names := []string{"apt", "dnf", "brew", "winget", "choco"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			pm, err := ByName(name)
			if err != nil {
				t.Fatalf("ByName(%q): %v", name, err)
			}
			if pm.Name() != name {
				t.Errorf("pm.Name() = %q, want %q", pm.Name(), name)
			}
		})
	}
}

func TestByName_External(t *testing.T) {
	_, err := ByName("external")
	if err == nil {
		t.Error("expected error for external package manager")
	}
	if !strings.Contains(err.Error(), "registry add") {
		t.Errorf("error should mention registry add, got: %v", err)
	}
	if !strings.Contains(err.Error(), "lifecycle") {
		t.Errorf("error should mention lifecycle, got: %v", err)
	}
}

func TestByName_Unknown(t *testing.T) {
	_, err := ByName("pacman")
	if err == nil {
		t.Error("expected error for unknown package manager")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("error should mention unknown, got: %v", err)
	}
}
