// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"testing"
)

func TestBuildClientConfigLoadingRulesWithExplicitKubeconfig(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfigPath := filepath.Join(tmpDir, "config")
	t.Setenv("KUBECONFIG", "old-config")

	if err := os.WriteFile(kubeconfigPath, []byte("apiVersion: v1\nkind: Config\n"), 0600); err != nil {
		t.Fatalf("failed to write temporary kubeconfig: %v", err)
	}

	rules, err := buildClientConfigLoadingRules(kubeconfigPath)
	if err != nil {
		t.Fatalf("buildClientConfigLoadingRules returned error: %v", err)
	}
	if rules == nil {
		t.Fatal("buildClientConfigLoadingRules returned nil rules")
	}
	if rules.ExplicitPath != kubeconfigPath {
		t.Fatalf("expected ExplicitPath %q, got %q", kubeconfigPath, rules.ExplicitPath)
	}
	if got := os.Getenv("KUBECONFIG"); got != "old-config" {
		t.Fatalf("expected KUBECONFIG to stay %q, got %q", "old-config", got)
	}
}

func TestInitKubeClientSetsExplicitKubeconfigEnv(t *testing.T) {
	kubeconfigPath := filepath.Join(t.TempDir(), "config")
	t.Setenv("KUBECONFIG", "old-config")
	kubeconfig := []byte(`apiVersion: v1
kind: Config
clusters:
- name: test
  cluster:
    server: https://127.0.0.1
contexts:
- name: test
  context:
    cluster: test
    user: test
current-context: test
users:
- name: test
  user:
    token: test
`)
	if err := os.WriteFile(kubeconfigPath, kubeconfig, 0600); err != nil {
		t.Fatalf("failed to write temporary kubeconfig: %v", err)
	}

	if _, _, _, err := initKubeClient(kubeconfigPath); err != nil {
		t.Fatalf("initKubeClient returned error: %v", err)
	}
	if got := os.Getenv("KUBECONFIG"); got != kubeconfigPath {
		t.Fatalf("expected KUBECONFIG %q, got %q", kubeconfigPath, got)
	}
}

func TestBuildClientConfigLoadingRulesWithMultipleKubeconfigEnv(t *testing.T) {
	tmpDir := t.TempDir()
	kubeconfig1 := filepath.Join(tmpDir, "config1")
	kubeconfig2 := filepath.Join(tmpDir, "config2")

	for _, path := range []string{kubeconfig1, kubeconfig2} {
		if err := os.WriteFile(path, []byte("apiVersion: v1\nkind: Config\n"), 0600); err != nil {
			t.Fatalf("failed to write temporary kubeconfig %s: %v", path, err)
		}
	}

	kubeconfigEnv := kubeconfig1 + string(os.PathListSeparator) + kubeconfig2
	t.Setenv("KUBECONFIG", kubeconfigEnv)

	rules, err := buildClientConfigLoadingRules("")
	if err != nil {
		t.Fatalf("buildClientConfigLoadingRules returned error: %v", err)
	}
	if rules == nil {
		t.Fatal("buildClientConfigLoadingRules returned nil rules")
	}
	if rules.ExplicitPath != "" {
		t.Fatalf("expected ExplicitPath to be empty when kubeconfig is not explicitly provided, got %q", rules.ExplicitPath)
	}
	expectedPrecedence := []string{kubeconfig1, kubeconfig2}
	if !reflect.DeepEqual(rules.Precedence, expectedPrecedence) {
		t.Fatalf("expected Precedence %v, got %v", expectedPrecedence, rules.Precedence)
	}
	if got := os.Getenv("KUBECONFIG"); got != kubeconfigEnv {
		t.Fatalf("expected KUBECONFIG to stay %q, got %q", kubeconfigEnv, got)
	}
}

func TestBuildClientConfigLoadingRulesWithMissingExplicitKubeconfig(t *testing.T) {
	kubeconfigPath := filepath.Join(t.TempDir(), "missing-config")

	rules, err := buildClientConfigLoadingRules(kubeconfigPath)
	if err == nil {
		t.Fatalf("expected an error for missing kubeconfig %q", kubeconfigPath)
	}
	if rules != nil {
		t.Fatalf("expected nil rules, got %#v", rules)
	}
}

func TestBuildClientConfigLoadingRulesWithoutKubeconfig(t *testing.T) {
	t.Setenv("KUBECONFIG", "")

	rules, err := buildClientConfigLoadingRules("")
	if err != nil {
		t.Fatalf("buildClientConfigLoadingRules returned error: %v", err)
	}
	if rules == nil {
		t.Fatal("buildClientConfigLoadingRules returned nil rules")
	}
	if rules.ExplicitPath != "" {
		t.Fatalf("expected ExplicitPath to be empty, got %q", rules.ExplicitPath)
	}
	if got := os.Getenv("KUBECONFIG"); got != "" {
		t.Fatalf("expected KUBECONFIG to stay empty, got %q", got)
	}
}

func TestSetupExplicitKubeconfigExpandsHome(t *testing.T) {
	currentUser, err := user.Current()
	if err != nil {
		t.Fatalf("failed to get current user: %v", err)
	}

	tests := []struct {
		name       string
		kubeconfig string
	}{
		{name: "home with trailing separator", kubeconfig: "~/"},
		{name: "bare home", kubeconfig: "~"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("KUBECONFIG", "old-config")

			got, err := setupExplicitKubeconfig(tt.kubeconfig)
			if err != nil {
				t.Fatalf("setupExplicitKubeconfig returned error: %v", err)
			}
			if got != currentUser.HomeDir {
				t.Fatalf("expected expanded path %q, got %q", currentUser.HomeDir, got)
			}
			if env := os.Getenv("KUBECONFIG"); env != "old-config" {
				t.Fatalf("expected KUBECONFIG to stay %q, got %q", "old-config", env)
			}
		})
	}
}
