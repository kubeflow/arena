// Copyright 2024 The Kubeflow Authors.
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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessSetFileOptions(t *testing.T) {
	dir := t.TempDir()

	fooPath := filepath.Join(dir, "foo.txt")
	if err := os.WriteFile(fooPath, []byte("foo-content"), 0o644); err != nil {
		t.Fatalf("write foo: %v", err)
	}

	barPath := filepath.Join(dir, "bar.txt")
	if err := os.WriteFile(barPath, []byte("bar-content"), 0o644); err != nil {
		t.Fatalf("write bar: %v", err)
	}

	t.Run("happy path single nested key", func(t *testing.T) {
		vals := map[string]interface{}{}
		opts := []string{fmt.Sprintf("--set-file configFiles.abc.content=%s", fooPath)}
		if err := processSetFileOptions(vals, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got, ok := getNestedString(vals, "configFiles", "abc", "content")
		if !ok {
			t.Fatalf("value not set; got %#v", vals)
		}
		if got != "foo-content" {
			t.Fatalf("expected foo-content, got %q", got)
		}
	})

	t.Run("nested keys merge into existing map without data loss", func(t *testing.T) {
		vals := map[string]interface{}{
			"configFiles": map[string]interface{}{
				"existing": map[string]interface{}{
					"keep": "keep-val",
				},
			},
		}
		opts := []string{
			fmt.Sprintf("--set-file configFiles.existing.content=%s", fooPath),
			fmt.Sprintf("--set-file configFiles.other.content=%s", barPath),
		}
		if err := processSetFileOptions(vals, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if v, ok := getNestedString(vals, "configFiles", "existing", "keep"); !ok || v != "keep-val" {
			t.Fatalf("expected pre-existing 'keep' value to be preserved; got %#v", vals)
		}
		if v, ok := getNestedString(vals, "configFiles", "existing", "content"); !ok || v != "foo-content" {
			t.Fatalf("expected new content for 'existing'; got %#v", vals)
		}
		if v, ok := getNestedString(vals, "configFiles", "other", "content"); !ok || v != "bar-content" {
			t.Fatalf("expected content for 'other'; got %#v", vals)
		}
	})

	t.Run("ignores non --set-file options", func(t *testing.T) {
		vals := map[string]interface{}{}
		opts := []string{
			"--set foo=bar",
			fmt.Sprintf("--set-file configFiles.abc.content=%s", fooPath),
			"--namespace default",
			"--debug",
		}
		if err := processSetFileOptions(vals, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, exists := vals["foo"]; exists {
			t.Fatalf("--set option must not be processed as a set-file; got %#v", vals)
		}
		if _, exists := vals["namespace"]; exists {
			t.Fatalf("--namespace option must not be processed; got %#v", vals)
		}
		if v, ok := getNestedString(vals, "configFiles", "abc", "content"); !ok || v != "foo-content" {
			t.Fatalf("expected set-file content to be applied; got %#v", vals)
		}
	})

	t.Run("supports --set-file=key=path form", func(t *testing.T) {
		vals := map[string]interface{}{}
		opts := []string{fmt.Sprintf("--set-file=configFiles.abc.content=%s", fooPath)}
		if err := processSetFileOptions(vals, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v, ok := getNestedString(vals, "configFiles", "abc", "content"); !ok || v != "foo-content" {
			t.Fatalf("expected content via equals-form flag; got %#v", vals)
		}
	})

	t.Run("missing file returns wrapped error", func(t *testing.T) {
		vals := map[string]interface{}{}
		missing := filepath.Join(dir, "does-not-exist.txt")
		opts := []string{fmt.Sprintf("--set-file key=%s", missing)}
		err := processSetFileOptions(vals, opts)
		if err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
		if !strings.Contains(err.Error(), "failed to read") {
			t.Fatalf("expected error to mention read failure; got %v", err)
		}
	})

	t.Run("malformed option missing equals returns error", func(t *testing.T) {
		vals := map[string]interface{}{}
		opts := []string{"--set-file no-equals-sign"}
		if err := processSetFileOptions(vals, opts); err == nil {
			t.Fatal("expected error for malformed option, got nil")
		}
	})

	t.Run("empty key returns error", func(t *testing.T) {
		vals := map[string]interface{}{}
		opts := []string{fmt.Sprintf("--set-file =%s", fooPath)}
		if err := processSetFileOptions(vals, opts); err == nil {
			t.Fatal("expected error for empty key, got nil")
		}
	})

	t.Run("intermediate non-map value returns error instead of overwriting", func(t *testing.T) {
		vals := map[string]interface{}{
			"configFiles": "already-a-string",
		}
		opts := []string{fmt.Sprintf("--set-file configFiles.abc.content=%s", fooPath)}
		err := processSetFileOptions(vals, opts)
		if err == nil {
			t.Fatal("expected error for non-map intermediate key, got nil")
		}
		// Verify the pre-existing scalar was NOT overwritten.
		if s, ok := vals["configFiles"].(string); !ok || s != "already-a-string" {
			t.Fatalf("pre-existing scalar should be preserved on error; got %#v", vals["configFiles"])
		}
	})

	t.Run("empty segment in dotted key returns error", func(t *testing.T) {
		vals := map[string]interface{}{}
		opts := []string{fmt.Sprintf("--set-file configFiles..content=%s", fooPath)}
		if err := processSetFileOptions(vals, opts); err == nil {
			t.Fatal("expected error for empty key segment, got nil")
		}
	})
}

// getNestedString is a small test helper that walks a nested
// map[string]interface{} chain and returns the terminal string value.
func getNestedString(vals map[string]interface{}, parts ...string) (string, bool) {
	current := vals
	for i, p := range parts {
		v, ok := current[p]
		if !ok {
			return "", false
		}
		if i == len(parts)-1 {
			s, ok := v.(string)
			return s, ok
		}
		m, ok := v.(map[string]interface{})
		if !ok {
			return "", false
		}
		current = m
	}
	return "", false
}
