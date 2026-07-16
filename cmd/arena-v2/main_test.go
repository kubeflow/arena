package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatError_BasicError(t *testing.T) {
	err := errors.New("simple error")

	result := formatError(err, false)
	assert.Equal(t, "Error: simple error\n", result)
}

func TestFormatError_WrappedError(t *testing.T) {
	innerErr := errors.New("connection refused")
	middleErr := fmt.Errorf("failed to connect: %w", innerErr)
	outerErr := fmt.Errorf("failed to create client: %w", middleErr)

	result := formatError(outerErr, false)
	assert.Contains(t, result, "Error: failed to create client")
	assert.Contains(t, result, "failed to connect")
	assert.Contains(t, result, "connection refused")
}

func TestFormatError_DebugMode(t *testing.T) {
	err := errors.New("test error")

	result := formatError(err, true)
	assert.Contains(t, result, "Error: test error")
	assert.Contains(t, result, "Full error chain:")
}
