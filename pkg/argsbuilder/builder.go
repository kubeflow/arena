package argsbuilder

import (
	"github.com/spf13/cobra"
)

// ArgsBuilder
type ArgsBuilder interface {
	AddSubBuilder(b ...ArgsBuilder) ArgsBuilder
	PreBuild() error
	Build() error
	AddCommandFlags(command *cobra.Command)
	GetName() string
	AddArgValue(key string, value interface{}) ArgsBuilder
}
