package cli

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubeflow/arena/pkg/provider"
	"github.com/kubeflow/arena/pkg/task"
)

// allExamplesDir resolves examples/v2 relative to pkg/cli/. Unlike examplesDir,
// which is scoped to the quickstart subset used by the CRD-shape tests, this
// covers every shipped example so none of them can rot unnoticed.
func allExamplesDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Join(wd, "..", "..", "examples", "v2")
}

// TestAllExamplesBuildCRD walks examples/v2 recursively and checks that every
// example still parses, validates, and produces a CRD through its framework's
// provider. The examples are user-facing documentation, so a schema change that
// invalidates one should fail here rather than in someone's terminal.
//
// Scope: this catches parse errors, Validate() failures and BuildCRD failures.
// It does not catch misspelled field names, because the YAML decoder ignores
// unknown keys -- a typo silently drops the value and only surfaces if the
// zero value happens to fail validation. Strict decoding would be needed for
// that, which is a larger change than this test.
func TestAllExamplesBuildCRD(t *testing.T) {
	root := allExamplesDir(t)

	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if ext := strings.ToLower(filepath.Ext(path)); ext == ".yaml" || ext == ".yml" {
			files = append(files, path)
		}
		return nil
	})
	require.NoError(t, err, "walking %s", root)
	require.NotEmpty(t, files, "no example YAML found under %s", root)

	for _, path := range files {
		rel, relErr := filepath.Rel(root, path)
		require.NoError(t, relErr)

		t.Run(rel, func(t *testing.T) {
			taskObj, err := task.LoadFromFile(path)
			require.NoError(t, err, "example failed to load")

			p, err := getProvider(taskObj.Framework.Name)
			require.NoError(t, err, "no provider for framework %q", taskObj.Framework.Name)

			// MPIJob's stored API version is normally read from the cluster; pin
			// it so the check stays offline.
			if mpiP, ok := p.(*provider.MPIProvider); ok {
				mpiP.APIVersion = "v1"
			}

			crd, err := p.BuildCRD(taskObj)
			require.NoError(t, err, "provider failed to build a CRD")
			assert.NotEmpty(t, crd.GetKind(), "CRD has no kind")
			assert.Equal(t, taskObj.Name, crd.GetName(), "CRD name should match the task name")
		})
	}
}
