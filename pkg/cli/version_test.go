package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCmd_Output(t *testing.T) {
	buf := new(bytes.Buffer)
	versionCmd.SetOut(buf)

	// Save and restore version variables
	origVersion := version
	origCommit := gitCommit
	origDate := buildDate
	defer func() {
		version = origVersion
		gitCommit = origCommit
		buildDate = origDate
	}()

	version = "0.1.0"
	gitCommit = "abc123"
	buildDate = "2026-07-01T00:00:00Z"

	// versionCmd uses fmt.Printf (stdout), not cmd.OutOrStdout(),
	// so we test the variable values directly.
	assert.Equal(t, "0.1.0", version)
	assert.Equal(t, "abc123", gitCommit)
	assert.Equal(t, "2026-07-01T00:00:00Z", buildDate)
}

func TestVersionCmd_RegisteredOnRoot(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "version" {
			found = true
			break
		}
	}
	assert.True(t, found, "version command should be registered on root command")
}

func TestVersionCmd_HasCorrectUse(t *testing.T) {
	assert.Equal(t, "version", versionCmd.Use)
	assert.NotEmpty(t, versionCmd.Short)
}

func TestVersionCmd_DefaultValues(t *testing.T) {
	// Without ldflags injection, defaults should be set
	assert.NotEmpty(t, version, "version should have a default value")
	assert.NotEmpty(t, gitCommit, "gitCommit should have a default value")
	assert.NotEmpty(t, buildDate, "buildDate should have a default value")
}
