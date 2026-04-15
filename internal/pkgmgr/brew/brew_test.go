package brew

import (
	"reflect"
	"testing"
)

func TestManager_Name(t *testing.T) {
	m := &Manager{}
	if got := m.Name(); got != "brew" {
		t.Errorf("Name() = %q, want %q", got, "brew")
	}
}

func TestManager_InstallCommand(t *testing.T) {
	m := &Manager{}
	tests := []struct {
		name string
		pkgs []string
		want []string
	}{
		{
			name: "single package",
			pkgs: []string{"postgresql@16"},
			want: []string{"brew", "install", "postgresql@16"},
		},
		{
			name: "multiple packages",
			pkgs: []string{"postgresql@16", "libpq"},
			want: []string{"brew", "install", "postgresql@16", "libpq"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.InstallCommand(tt.pkgs)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InstallCommand(%v) = %v, want %v", tt.pkgs, got, tt.want)
			}
		})
	}
}

func TestManager_UninstallCommand(t *testing.T) {
	m := &Manager{}
	got := m.UninstallCommand([]string{"postgresql@16"})
	want := []string{"brew", "uninstall", "postgresql@16"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UninstallCommand = %v, want %v", got, want)
	}
}

func TestManager_ServiceStart(t *testing.T) {
	m := &Manager{}
	got := m.ServiceStart("postgresql@16")
	want := []string{"brew", "services", "start", "postgresql@16"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ServiceStart = %v, want %v", got, want)
	}
}

func TestManager_ServiceStop(t *testing.T) {
	m := &Manager{}
	got := m.ServiceStop("postgresql@16")
	want := []string{"brew", "services", "stop", "postgresql@16"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ServiceStop = %v, want %v", got, want)
	}
}

func TestManager_NoSudo(t *testing.T) {
	m := &Manager{}
	cmd := m.InstallCommand([]string{"postgresql@16"})
	if cmd[0] == "sudo" {
		t.Error("brew commands must NOT use sudo")
	}
}
