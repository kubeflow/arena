package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogsCmd_ArgsValidator(t *testing.T) {
	assert.Error(t, logsCmd.Args(logsCmd, []string{}))
	assert.NoError(t, logsCmd.Args(logsCmd, []string{"my-job"}))
	assert.Error(t, logsCmd.Args(logsCmd, []string{"a", "b"}))
}

func TestLogsCmd_FollowFlag(t *testing.T) {
	f := logsCmd.Flags().Lookup("follow")
	require.NotNil(t, f, "expected --follow flag to be registered")
	assert.Equal(t, "f", f.Shorthand)
	assert.Equal(t, "false", f.DefValue)
}

func TestLogsCmd_TailFlag(t *testing.T) {
	f := logsCmd.Flags().Lookup("tail")
	require.NotNil(t, f, "expected --tail flag to be registered")
	assert.Equal(t, "-1", f.DefValue)
}

func TestLogsCmd_RegisteredOnJob(t *testing.T) {
	found := false
	for _, cmd := range jobCmd.Commands() {
		if cmd.Use == "logs <name>" {
			found = true
			break
		}
	}
	assert.True(t, found, "logs command should be registered on job command")
}

func TestLogsCmd_RunE_FailsWithInvalidKubeconfig(t *testing.T) {
	orig := kubeconfig
	defer func() { kubeconfig = orig }()

	kubeconfig = "/nonexistent/kubeconfig"
	err := logsCmd.RunE(logsCmd, []string{"my-job"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to build config")
}

func TestLogsCmd_HasCorrectUse(t *testing.T) {
	assert.Equal(t, "logs <name>", logsCmd.Use)
	assert.NotEmpty(t, logsCmd.Short)
}

func TestLogsCmd_RunE_RequiresKubeconfig(t *testing.T) {
	orig := kubeconfig
	defer func() { kubeconfig = orig }()

	t.Setenv("KUBECONFIG", "/nonexistent/env-kubeconfig")
	kubeconfig = ""
	err := logsCmd.RunE(logsCmd, []string{"my-job"})
	// Without a valid kubeconfig, client creation or REST config should fail.
	assert.Error(t, err)
}

func TestLogsCmd_PodFlag(t *testing.T) {
	f := logsCmd.Flags().Lookup("pod")
	require.NotNil(t, f, "expected --pod flag to be registered")
	assert.Equal(t, "", f.DefValue)
}

func TestLogsCmd_ContainerFlag(t *testing.T) {
	f := logsCmd.Flags().Lookup("container")
	require.NotNil(t, f, "expected --container flag to be registered")
	assert.Equal(t, "", f.DefValue)
}
