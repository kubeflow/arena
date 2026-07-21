// Package cli - integration tests
//go:build integration
// +build integration

package cli

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubeflow/arena/pkg/client"
	outputpkg "github.com/kubeflow/arena/pkg/output"
)

// Integration tests require a real Kubernetes cluster.
// Run with: go test -tags integration -v ./pkg/cli/ -run Integration
// KUBECONFIG must be set (defaults to ~/pro5000)

func integrationKubeconfig() string {
	kc := os.Getenv("KUBECONFIG")
	if kc == "" {
		home, _ := os.UserHomeDir()
		kc = home + "/pro5000"
	}
	return kc
}

func TestIntegration_ListJobs(t *testing.T) {
	k8sClient, err := client.NewClient(integrationKubeconfig(), "")
	require.NoError(t, err)

	ctx := context.Background()
	jobs, err := k8sClient.List(ctx, "PyTorchJob", "default", "")
	require.NoError(t, err)
	assert.NotEmpty(t, jobs, "expected at least one PyTorchJob in default namespace")

	for _, job := range jobs {
		status := extractJobPhase(job)
		assert.NotEqual(t, "", status, "job %s should have a status", job.GetName())
		assert.NotEqual(t, "Created", status,
			"job %s should not show Created as final status (reverse-scan regression)", job.GetName())
	}
}

func TestIntegration_GetJob(t *testing.T) {
	k8sClient, err := client.NewClient(integrationKubeconfig(), "")
	require.NoError(t, err)

	ctx := context.Background()
	jobs, err := k8sClient.List(ctx, "PyTorchJob", "default", "")
	require.NoError(t, err)
	require.NotEmpty(t, jobs, "need at least one job to test get")

	name := jobs[0].GetName()
	job, err := k8sClient.Get(ctx, "PyTorchJob", "default", name)
	require.NoError(t, err)
	assert.Equal(t, name, job.GetName())

	info := extractJobStatus(job, "PyTorchJob")
	assert.Equal(t, name, info.Name)
	assert.Equal(t, "default", info.Namespace)
	assert.NotEmpty(t, info.Status)
	assert.NotEqual(t, "Created", info.Status)
}

func TestIntegration_CheckCRDs(t *testing.T) {
	k8sClient, err := client.NewClient(integrationKubeconfig(), "")
	require.NoError(t, err)

	ctx := context.Background()

	_, err = k8sClient.List(ctx, "PyTorchJob", "default", "")
	require.NoError(t, err, "PyTorchJob CRD should be installed")

	_, err = k8sClient.List(ctx, "TFJob", "default", "")
	require.NoError(t, err, "TFJob CRD should be installed")
}

func TestIntegration_DetectJobTypeForV2Job(t *testing.T) {
	// This test verifies detectJobType works for v2-created jobs
	// (those with ConfigMap anchors). It will be skipped if no v2 job exists.
	k8sClient, err := client.NewClient(integrationKubeconfig(), "")
	require.NoError(t, err)

	ctx := context.Background()

	// Try to find a v2 job (one with a ConfigMap anchor)
	jobs, err := k8sClient.List(ctx, "PyTorchJob", "default", "")
	require.NoError(t, err)

	for _, job := range jobs {
		name := job.GetName()
		_, err := k8sClient.Get(ctx, "ConfigMap", "default", name)
		if err == nil {
			// Found a v2 job, test detectJobType
			kind, err := detectJobType(ctx, k8sClient, "default", name)
			require.NoError(t, err)
			assert.Equal(t, "PyTorchJob", kind)
			return
		}
	}

	t.Skip("no v2-created jobs found in cluster (jobs without ConfigMap anchors are v1-created)")
}

func TestIntegration_StatusExtraction(t *testing.T) {
	k8sClient, err := client.NewClient(integrationKubeconfig(), "")
	require.NoError(t, err)

	ctx := context.Background()
	jobs, err := k8sClient.List(ctx, "PyTorchJob", "default", "")
	require.NoError(t, err)

	for _, job := range jobs {
		status := extractJobStatus(job, "PyTorchJob")
		assert.NotEmpty(t, status.Status, "job %s must have status", job.GetName())
		assert.NotEqual(t, "Created", status.Status,
			"reverse-scan bug: job %s shows Created instead of actual status", job.GetName())

		// Verify replicas/ready are non-negative
		assert.GreaterOrEqual(t, status.Replicas, 0)
		assert.GreaterOrEqual(t, status.Ready, 0)
	}
}

// TestIntegration_TopJobDryRun verifies that the "top job" command parses
// correctly and reaches the point where it attempts to create a Kubernetes
// client. Without a live cluster this is expected to fail with a client
// creation error — any other error indicates a regression in command wiring
// (e.g. flag registration or format validation).
func TestIntegration_TopJobDryRun(t *testing.T) {
	// Use rootCmd.SetArgs with the full command path so cobra traverses into
	// "top job" rather than falling back to the root help screen. Pass a
	// nonexistent kubeconfig so client creation fails fast instead of timing
	// out against an unreachable cluster.
	rootCmd.SetArgs([]string{"top", "job", "-o", string(outputpkg.FormatJSON), "--kubeconfig", "/nonexistent/arena-test-kubeconfig"})
	defer rootCmd.SetArgs([]string{})

	err := rootCmd.Execute()
	// Without a cluster we expect a K8s client error; anything else is a bug
	// in command wiring (flag registration, format validation, etc.).
	require.Error(t, err, "top job should fail without a reachable cluster")
	assert.True(t, strings.Contains(err.Error(), "failed to create K8s client"),
		"expected K8s client error, got: %v", err)
}

// TestIntegration_ListOutputFormats verifies that the "job list" command
// accepts all four supported output formats (table, wide, json, yaml). Each
// format should pass flag validation and reach the cluster-connection step;
// a format-specific error (e.g. "invalid output format") indicates a
// regression in the -o persistent flag wiring.
func TestIntegration_ListOutputFormats(t *testing.T) {
	formats := []string{string(outputpkg.FormatTable), string(outputpkg.FormatWide), string(outputpkg.FormatJSON), string(outputpkg.FormatYAML)}
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			// Use the -o flag on "job list" via rootCmd so cobra traverses
			// the full command tree and exercises the persistent flag wiring
			// on jobCmd. Use a nonexistent kubeconfig for fast failure.
			rootCmd.SetArgs([]string{"job", "list", "-o", format, "--kubeconfig", "/nonexistent/arena-test-kubeconfig"})
			defer rootCmd.SetArgs([]string{})

			err := rootCmd.Execute()
			// The command must fail (no cluster), but the failure must be a
			// K8s client error — not a format validation error. This proves
			// the -o persistent flag on jobCmd accepts every listed format.
			require.Error(t, err, "job list -o %s should fail without a reachable cluster", format)
			assert.True(t, strings.Contains(err.Error(), "failed to create K8s client"),
				"format %s: expected K8s client error, got: %v", format, err)
		})
	}
}
