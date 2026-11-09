package cmd

import (
	"github.com/spf13/cobra"
)

type statusOptions struct {
}

func newStatusCmd(global *globalOptions) *cobra.Command {
	opts := &initOptions{}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show this node's details",
		Long: `View up this machine as a GoCrypt node.

status shows the node's details, including the Hostname, Node ID,
Home directory, Storage directory and Key information.

Key information is safe to print and nothing sensitive is exposed.
If any problems are found this command lists them all.`,
		Example: "gocrypt status",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// From here on, errors are about the node, not about how the
			// command was typed, so don't print the usage text with them.
			cmd.SilenceUsage = true
			return runStatus(cmd, global, opts)
		},
	}

	return cmd
}

func runStatus(cmd *cobra.Command, global *globalOptions, opts *initOptions) error {
	return nil
}
