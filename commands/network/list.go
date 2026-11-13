package network

import (
	"github.com/its-kos/gocrypt/cli"
	"github.com/spf13/cobra"
)

func newNetworkListCmd(global *cli.GlobalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List networks this node belongs to",
		Long: `List all logical networks this node is a member of.

Displays information about each network, including its name and
network ID.

If this node does not belong to any networks, an empty list is shown.`,
		Example: `  gocrypt network list`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true
			return runNetworkList(cmd, global)
		},
	}
}

func runNetworkList(cmd *cobra.Command, global *cli.GlobalOptions) error {
	return nil
}
