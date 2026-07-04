// Copyright 2026 The Kubeflow Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessSetFileOptions(t *testing.T) {
	// Create a temp directory for test files
	tempDir := t.TempDir()
	file1Path := filepath.Join(tempDir, "config1.yaml")
	file2Path := filepath.Join(tempDir, "config2.yaml")

	content1 := "foo: bar\nhello: world"
	content2 := "some: content"

	if err := os.WriteFile(file1Path, []byte(content1), 0644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}
	if err := os.WriteFile(file2Path, []byte(content2), 0644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}

	t.Run("happy path", func(t *testing.T) {
		values := make(map[string]interface{})
		options := []string{
			"--set-file", "configFiles.hash.config-0.content=" + filepath.ToSlash(file1Path),
			"--set-file=configFiles.hash.config-1.content=" + filepath.ToSlash(file2Path),
			"--some-other-option",
		}

		err := processSetFileOptions(values, options)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify values
		configFiles, ok := values["configFiles"].(map[string]interface{})
		if !ok {
			t.Fatalf("configFiles is not map[string]interface{}")
		}
		hash, ok := configFiles["hash"].(map[string]interface{})
		if !ok {
			t.Fatalf("hash is not map[string]interface{}")
		}

		config0, ok := hash["config-0"].(map[string]interface{})
		if !ok {
			t.Fatalf("config-0 is not map[string]interface{}")
		}
		if config0["content"] != content1 {
			t.Errorf("expected content1, got %v", config0["content"])
		}

		config1, ok := hash["config-1"].(map[string]interface{})
		if !ok {
			t.Fatalf("config-1 is not map[string]interface{}")
		}
		if config1["content"] != content2 {
			t.Errorf("expected content2, got %v", config1["content"])
		}
	})

	t.Run("malformed option", func(t *testing.T) {
		values := make(map[string]interface{})
		options := []string{
			"--set-file", "invalidformat",
		}
		err := processSetFileOptions(values, options)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "has no value") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		values := make(map[string]interface{})
		options := []string{
			"--set-file", "key=" + filepath.ToSlash(filepath.Join(tempDir, "nonexistent.yaml")),
		}
		err := processSetFileOptions(values, options)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to read file") && !strings.Contains(err.Error(), "no such file or directory") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("empty key part", func(t *testing.T) {
		values := make(map[string]interface{})
		options := []string{
			"--set-file", ".=" + filepath.ToSlash(file1Path),
		}
		err := processSetFileOptions(values, options)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("overwrite existing scalar conflict", func(t *testing.T) {
		values := map[string]interface{}{
			"conflictKey": "scalarValue",
		}
		options := []string{
			"--set-file", "conflictKey.nested=" + filepath.ToSlash(file1Path),
		}
		err := processSetFileOptions(values, options)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "interface conversion") {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}
