package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCmd_HasDebugFlag(t *testing.T) {
	f := NewRootCommand().PersistentFlags().Lookup("debug")
	assert.NotNil(t, f, "debug flag should be registered")
	assert.Equal(t, "false", f.DefValue)
}

func TestRootCmd_HasVerboseFlag(t *testing.T) {
	f := NewRootCommand().PersistentFlags().Lookup("verbose")
	assert.NotNil(t, f, "verbose flag should be registered")
	assert.Equal(t, "0", f.DefValue)
}

func TestDebugMode_Variable(t *testing.T) {
	assert.False(t, debugMode, "debugMode should default to false")
}
