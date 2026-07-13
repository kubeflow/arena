package client

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// writeTempKubeconfig creates a minimal kubeconfig file for testing.
func writeTempKubeconfig(t *testing.T, contextName, namespace string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "kubeconfig")

	nsBlock := ""
	if namespace != "" {
		nsBlock = "    namespace: " + namespace + "\n"
	}

	content := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
` + nsBlock + `  name: ` + contextName + `
current-context: ` + contextName + `
users:
- name: test-user
  user:
    token: test-token
`
	err := os.WriteFile(path, []byte(content), 0600)
	require.NoError(t, err)
	return path
}

func TestLoadRestConfig_ExplicitPath(t *testing.T) {
	path := writeTempKubeconfig(t, "test-ctx", "")
	config, err := LoadRestConfig(path, "")
	require.NoError(t, err)
	require.NotNil(t, config)
	require.Equal(t, "https://127.0.0.1:6443", config.Host)
}

func TestLoadRestConfig_WithContext(t *testing.T) {
	path := writeTempKubeconfig(t, "my-context", "")
	config, err := LoadRestConfig(path, "my-context")
	require.NoError(t, err)
	require.NotNil(t, config)
	require.Equal(t, "https://127.0.0.1:6443", config.Host)
}

func TestLoadRestConfig_QPSBurst(t *testing.T) {
	path := writeTempKubeconfig(t, "test-ctx", "")
	config, err := LoadRestConfig(path, "")
	require.NoError(t, err)
	require.Equal(t, float32(10), config.QPS)
	require.Equal(t, 20, config.Burst)
}

func TestLoadRestConfig_BadPath(t *testing.T) {
	_, err := LoadRestConfig("/nonexistent/kubeconfig", "")
	require.Error(t, err)
}

func TestLoadRestConfig_BadContext(t *testing.T) {
	path := writeTempKubeconfig(t, "test-ctx", "")
	_, err := LoadRestConfig(path, "nonexistent-context")
	require.Error(t, err)
}

func TestResolveNamespace_CLIFlag(t *testing.T) {
	path := writeTempKubeconfig(t, "test-ctx", "kube-ns")
	// CLI flag takes priority
	ns := ResolveNamespace(path, "", "cli-ns")
	require.Equal(t, "cli-ns", ns)
}

func TestResolveNamespace_KubeconfigContext(t *testing.T) {
	path := writeTempKubeconfig(t, "test-ctx", "kube-ns")
	// No CLI flag, should use kubeconfig context namespace
	ns := ResolveNamespace(path, "", "")
	require.Equal(t, "kube-ns", ns)
}

func TestResolveNamespace_DefaultFallback(t *testing.T) {
	path := writeTempKubeconfig(t, "test-ctx", "")
	// No CLI flag, no namespace in context, should fall back to "default"
	ns := ResolveNamespace(path, "", "")
	require.Equal(t, "default", ns)
}

func TestResolveNamespace_BadPath_Fallback(t *testing.T) {
	// Bad kubeconfig path with no CLI flag should fall back to "default"
	ns := ResolveNamespace("/nonexistent/kubeconfig", "", "")
	require.Equal(t, "default", ns)
}

func TestResolveNamespace_CLIFlagOverridesContext(t *testing.T) {
	path := writeTempKubeconfig(t, "test-ctx", "from-context")
	ns := ResolveNamespace(path, "test-ctx", "from-cli")
	require.Equal(t, "from-cli", ns)
}
