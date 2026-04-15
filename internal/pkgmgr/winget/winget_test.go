package winget

import (
	"reflect"
	"testing"
)

func TestManager_Name(t *testing.T) {
	m := &Manager{}
	if got := m.Name(); got != "winget" {
		t.Errorf("Name() = %q, want %q", got, "winget")
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
			pkgs: []string{"PostgreSQL.PostgreSQL.16"},
			want: []string{"winget", "install", "--accept-package-agreements", "--accept-source-agreements", "--id", "PostgreSQL.PostgreSQL.16"},
		},
		{
			name: "multiple packages",
			pkgs: []string{"PostgreSQL.PostgreSQL.16", "PostgreSQL.pgAdmin"},
			want: []string{"winget", "install", "--accept-package-agreements", "--accept-source-agreements", "--id", "PostgreSQL.PostgreSQL.16", "--id", "PostgreSQL.pgAdmin"},
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
	got := m.UninstallCommand([]string{"PostgreSQL.PostgreSQL.16"})
	want := []string{"winget", "uninstall", "--id", "PostgreSQL.PostgreSQL.16"}
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
