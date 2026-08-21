package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCmd_Output(t *testing.T) {
	// Save and restore version variables
	origVersion := version
	origCommit := gitCommit
	origDate := buildDate
	origTag := gitTag
	origTreeState := gitTreeState
	defer func() {
		version = origVersion
		gitCommit = origCommit
		buildDate = origDate
		gitTag = origTag
		gitTreeState = origTreeState
	}()

	version = "0.1.0"
	gitCommit = "abc123"
	buildDate = "2026-07-01T00:00:00Z"
	gitTag = "v0.1.0"
	gitTreeState = "clean"

	cmd := newVersionCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	require.NoError(t, cmd.RunE(cmd, nil))

	assert.Contains(t, buf.String(), "Arena v2")
	assert.Contains(t, buf.String(), "Version:     0.1.0")
	assert.Contains(t, buf.String(), "Git Commit:  abc123")
	assert.Contains(t, buf.String(), "Build Date:  2026-07-01T00:00:00Z")
	assert.Contains(t, buf.String(), "Git Tag:     v0.1.0")
	assert.Contains(t, buf.String(), "Tree State:  clean")
}

func TestVersionCmd_RegisteredOnRoot(t *testing.T) {
	found := false
	for _, cmd := range NewRootCommand().Commands() {
		if cmd.Name() == "version" {
			found = true
			break
		}
	}
	assert.True(t, found, "version command should be registered on root command")
}

func TestVersionCmd_HasCorrectUse(t *testing.T) {
	cmd := newVersionCmd()
	assert.Equal(t, "version", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestVersionCmd_DefaultValues(t *testing.T) {
	// Without ldflags injection, defaults should be set
	assert.NotEmpty(t, version, "version should have a default value")
	assert.NotEmpty(t, gitCommit, "gitCommit should have a default value")
	assert.NotEmpty(t, buildDate, "buildDate should have a default value")
	assert.Empty(t, gitTag, "gitTag should default to empty")
	assert.Equal(t, "unknown", gitTreeState, "gitTreeState should default to unknown")
}
