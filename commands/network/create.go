package network

import (
	"github.com/its-kos/gocrypt/cli"
	"github.com/spf13/cobra"
)

func newNetworkCreateCmd(global *cli.GlobalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new network",
		Long: `Create a new logical network.

The current node becomes the initial member of the network.
A network ID is generated to uniquely identify the network.

Additional nodes can join the network using an invitation.`,
		Example: `  gocrypt network create homelab
  gocrypt network create family`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return runNetworkCreate(cmd, global, args[0])
		},
	}
}

func runNetworkCreate(cmd *cobra.Command, global *cli.GlobalOptions, name string) error {
	return nil
}
