package command

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

type CommandWrapper struct {
	runFunc (func(cmd *cobra.Command, args []string) error)
}

func NewCommandWrapper(run func(cmd *cobra.Command, args []string) error) *CommandWrapper {
	return &CommandWrapper{
		runFunc: run,
	}
}

func (wrapper *CommandWrapper) Run(cmd *cobra.Command, args []string) {
	err := wrapper.runFunc(cmd, args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
