package task

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadFromFile reads a YAML file and parses it into a validated Task.
func LoadFromFile(path string) (*Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", path, err)
	}
	return LoadFromBytes(data)
}

// LoadFromBytes parses YAML data into a Task and validates it.
func LoadFromBytes(data []byte) (*Task, error) {
	var t Task
	if err := yaml.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	t.SetDefaults()
	if err := Validate(&t); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	return &t, nil
}
