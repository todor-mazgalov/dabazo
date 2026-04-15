package dnf

import (
	"reflect"
	"testing"
)

func TestManager_Name(t *testing.T) {
	m := &Manager{}
	if got := m.Name(); got != "dnf" {
		t.Errorf("Name() = %q, want %q", got, "dnf")
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
			pkgs: []string{"postgresql16-server"},
			want: []string{"sudo", "dnf", "install", "-y", "postgresql16-server"},
		},
		{
			name: "multiple packages",
			pkgs: []string{"postgresql16-server", "postgresql16"},
			want: []string{"sudo", "dnf", "install", "-y", "postgresql16-server", "postgresql16"},
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
	got := m.UninstallCommand([]string{"postgresql16-server"})
	want := []string{"sudo", "dnf", "remove", "-y", "postgresql16-server"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UninstallCommand = %v, want %v", got, want)
	}
}

func TestManager_ServiceStart(t *testing.T) {
	m := &Manager{}
	got := m.ServiceStart("postgresql-16")
	want := []string{"sudo", "systemctl", "start", "postgresql-16"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ServiceStart = %v, want %v", got, want)
	}
}

func TestManager_ServiceStop(t *testing.T) {
	m := &Manager{}
	got := m.ServiceStop("postgresql-16")
	want := []string{"sudo", "systemctl", "stop", "postgresql-16"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ServiceStop = %v, want %v", got, want)
	}
}
