package cli

import (
	"context"

	"github.com/spf13/cobra"
)

// cmdContext returns the context associated with a cobra command.
// Falls back to context.Background() if the command has no context set
// (e.g., when RunE is called directly in tests).
func cmdContext(cmd *cobra.Command) context.Context {
	ctx := cmd.Context()
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
