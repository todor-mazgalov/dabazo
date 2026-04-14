package choco

import (
	"reflect"
	"testing"
)

func TestManager_Name(t *testing.T) {
	m := &Manager{}
	if got := m.Name(); got != "choco" {
		t.Errorf("Name() = %q, want %q", got, "choco")
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
			pkgs: []string{"postgresql"},
			want: []string{"choco", "install", "-y", "postgresql"},
		},
		{
			name: "multiple packages",
			pkgs: []string{"postgresql", "pgadmin4"},
			want: []string{"choco", "install", "-y", "postgresql", "pgadmin4"},
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
	got := m.UninstallCommand([]string{"postgresql"})
	want := []string{"choco", "uninstall", "-y", "postgresql"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UninstallCommand = %v, want %v", got, want)
	}
}

func TestManager_ServiceStart_UsesPgCtl(t *testing.T) {
	m := &Manager{}
	got := m.ServiceStart("/data/pg16")
	want := []string{"pg_ctl", "start", "-D", "/data/pg16"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ServiceStart = %v, want %v", got, want)
	}
}

func TestManager_ServiceStop_UsesPgCtl(t *testing.T) {
	m := &Manager{}
	got := m.ServiceStop("/data/pg16")
	want := []string{"pg_ctl", "stop", "-D", "/data/pg16"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ServiceStop = %v, want %v", got, want)
	}
}
