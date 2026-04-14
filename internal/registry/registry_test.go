// Package registry tests for instance registry operations.
package registry

import (
	"os"
	"path/filepath"
	"testing"

	"dabazo/internal/engines"
)

func setupTestDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	if os.Getenv("APPDATA") != "" {
		t.Setenv("APPDATA", dir)
	}
}

func TestLoadEmpty(t *testing.T) {
	setupTestDir(t)
	instances, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(instances) != 0 {
		t.Errorf("expected empty slice, got %d instances", len(instances))
	}
}

func TestAddAndFind(t *testing.T) {
	setupTestDir(t)
	inst := engines.Instance{
		Name:           "test",
		Engine:         "postgres",
		Version:        "16",
		Port:           5432,
		Host:           "localhost",
		PackageManager: "apt",
	}
	if err := Add(inst); err != nil {
		t.Fatalf("Add: %v", err)
	}

	found, err := Find("test")
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if found == nil {
		t.Fatal("expected instance, got nil")
	}
	if found.Name != "test" {
		t.Errorf("expected name 'test', got %q", found.Name)
	}
}

func TestAddDuplicate(t *testing.T) {
	setupTestDir(t)
	inst := engines.Instance{Name: "dup", Engine: "postgres", Version: "16", Port: 5432, Host: "localhost"}
	if err := Add(inst); err != nil {
		t.Fatalf("first Add: %v", err)
	}
	if err := Add(inst); err == nil {
		t.Error("expected error on duplicate add, got nil")
	}
}

func TestRemove(t *testing.T) {
	setupTestDir(t)
	inst := engines.Instance{Name: "rm", Engine: "postgres", Version: "16", Port: 5432, Host: "localhost"}
	if err := Add(inst); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := Remove("rm"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	found, err := Find("rm")
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if found != nil {
		t.Error("expected nil after removal")
	}
}

func TestRemoveNotFound(t *testing.T) {
	setupTestDir(t)
	if err := Remove("nonexistent"); err == nil {
		t.Error("expected error removing nonexistent instance")
	}
}

func TestResolve_SingleInstance(t *testing.T) {
	setupTestDir(t)
	inst := engines.Instance{Name: "only", Engine: "postgres", Version: "16", Port: 5432, Host: "localhost"}
	if err := Add(inst); err != nil {
		t.Fatalf("Add: %v", err)
	}

	// Without name - should resolve to the single instance.
	found, err := Resolve("")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if found.Name != "only" {
		t.Errorf("expected 'only', got %q", found.Name)
	}
}

func TestResolve_MultipleRequiresName(t *testing.T) {
	setupTestDir(t)
	_ = Add(engines.Instance{Name: "a", Engine: "postgres", Version: "16", Port: 5432, Host: "localhost"})
	_ = Add(engines.Instance{Name: "b", Engine: "postgres", Version: "17", Port: 5433, Host: "localhost"})

	_, err := Resolve("")
	if err == nil {
		t.Error("expected error when multiple instances and no name")
	}

	found, err := Resolve("b")
	if err != nil {
		t.Fatalf("Resolve with name: %v", err)
	}
	if found.Name != "b" {
		t.Errorf("expected 'b', got %q", found.Name)
	}
}

func TestResolve_Empty(t *testing.T) {
	setupTestDir(t)
	_, err := Resolve("")
	if err == nil {
		t.Error("expected error when no instances registered")
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	if os.Getenv("APPDATA") != "" {
		t.Setenv("APPDATA", dir)
	}

	inst := engines.Instance{Name: "dirtest", Engine: "postgres", Version: "16", Port: 5432, Host: "localhost"}
	if err := Save([]engines.Instance{inst}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	fp := filepath.Join(dir, "dabazo", "instances.json")
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		t.Error("expected instances.json to be created")
	}
}
