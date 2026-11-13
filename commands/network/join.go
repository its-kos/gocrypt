package network

import (
	"github.com/its-kos/gocrypt/cli"
	"github.com/spf13/cobra"
)

func newNetworkJoinCmd(global *cli.GlobalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "join <invitation>",
		Short: "Join an existing network",
		Long: `Join an existing GoCrypt network using an invitation.

The invitation contains the information needed to identify the
network and establish membership.

The invitation must be valid and intended for this node.`,
		Example: `  gocrypt network join <invitation>`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return runNetworkJoin(cmd, global, args[0])
		},
	}
}

func runNetworkJoin(cmd *cobra.Command, global *cli.GlobalOptions, invitation string) error {
	return nil
}
