package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newSubmitFrameworkSubcommand builds the per-framework submit subcommand
// (submit pytorchjob, submit tfjob, …) for a supported registry entry. Each
// subcommand carries only the common flags plus its framework-specific
// subset, so --help stays focused and flags foreign to the framework fail
// with "unknown flag". Anything that does not match a subcommand name or
// alias (unknown types, v1-only types, case variants) falls back to the
// parent submit command unchanged.
func newSubmitFrameworkSubcommand(def frameworkDef) *cobra.Command {
	aliases := make([]string, 0, len(def.aliases))
	for _, a := range def.aliases {
		if a != def.cmdName {
			aliases = append(aliases, a)
		}
	}
	cmd := &cobra.Command{
		Use:     def.cmdName,
		Aliases: aliases,
		Short:   fmt.Sprintf("Submit a %s training job", def.original),
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// The label uses originalFramework(typed alias) exactly like the
			// parent path, so submit tf and submit tfjob both label
			// "tensorflow" (the canonical name, not the typed alias).
			return runSubmit(cmd, def.canonical, originalFramework(cmd.CalledAs()), args)
		},
	}
	registerSubmitCommonFlags(cmd)
	// Required only on subcommands: the parent's copy stays unmarked so a
	// bare `submit` reaches RunE and prints help instead of failing cobra's
	// required-flag validation; the parent path defers missing --name/--image
	// to task validation.
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("image")
	registerSubmitFrameworkFlags(cmd, def.canonical)
	registerSubmitCompatFlags(cmd)
	return cmd
}

func init() {
	for _, def := range frameworkRegistry {
		if def.canonical == "" {
			continue
		}
		submitCmd.AddCommand(newSubmitFrameworkSubcommand(def))
	}
}
