// Package registry manages the local instance registry stored in instances.json.
package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/todor-mazgalov/dabazo/internal/engines"
)

// configDir returns the OS-specific directory for dabazo configuration.
func configDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolving config directory: %w", err)
	}
	return filepath.Join(base, "dabazo"), nil
}

// filePath returns the full path to instances.json.
func filePath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "instances.json"), nil
}

// Load reads all instances from the registry file.
// Returns an empty slice if the file does not exist.
func Load() ([]engines.Instance, error) {
	fp, err := filePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading registry: %w", err)
	}
	var instances []engines.Instance
	if err := json.Unmarshal(data, &instances); err != nil {
		return nil, fmt.Errorf("parsing registry: %w", err)
	}
	return instances, nil
}

// Save writes the instances slice to the registry file.
func Save(instances []engines.Instance) error {
	fp, err := filePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(fp)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	data, err := json.MarshalIndent(instances, "", "  ")
	if err != nil {
		return fmt.Errorf("serializing registry: %w", err)
	}
	if err := os.WriteFile(fp, data, 0o644); err != nil {
		return fmt.Errorf("writing registry: %w", err)
	}
	return nil
}

// Add appends an instance to the registry. Returns an error if the name is taken.
func Add(inst engines.Instance) error {
	instances, err := Load()
	if err != nil {
		return err
	}
	for _, existing := range instances {
		if existing.Name == inst.Name {
			return fmt.Errorf("instance %q already exists in registry", inst.Name)
		}
	}
	instances = append(instances, inst)
	return Save(instances)
}

// Remove deletes an instance by name. Returns an error if not found.
func Remove(name string) error {
	instances, err := Load()
	if err != nil {
		return err
	}
	idx := -1
	for i, inst := range instances {
		if inst.Name == name {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("instance %q not found in registry", name)
	}
	instances = append(instances[:idx], instances[idx+1:]...)
	return Save(instances)
}

// Find looks up an instance by name. Returns nil if not found.
func Find(name string) (*engines.Instance, error) {
	instances, err := Load()
	if err != nil {
		return nil, err
	}
	for i := range instances {
		if instances[i].Name == name {
			return &instances[i], nil
		}
	}
	return nil, nil
}

// Resolve applies the --name resolution rule:
//   - 0 instances: error
//   - 1 instance: use it (name optional)
//   - >1 instances: name required
func Resolve(name string) (*engines.Instance, error) {
	instances, err := Load()
	if err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances registered; run `dabazo install` first")
	}
	if name != "" {
		for i := range instances {
			if instances[i].Name == name {
				return &instances[i], nil
			}
		}
		return nil, fmt.Errorf("instance %q not found in registry", name)
	}
	if len(instances) == 1 {
		return &instances[0], nil
	}
	names := make([]string, len(instances))
	for i, inst := range instances {
		names[i] = inst.Name
	}
	return nil, fmt.Errorf("multiple instances registered; specify --name (available: %v)", names)
}
