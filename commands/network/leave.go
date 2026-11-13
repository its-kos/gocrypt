package network

import (
	"github.com/its-kos/gocrypt/cli"
	"github.com/spf13/cobra"
)

func newNetworkLeaveCmd(global *cli.GlobalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "leave <network>",
		Short: "Leave a network",
		Long: `Remove this node's membership in a network.

This node will no longer participate in the specified network.

Leaving a network does not necessarily delete the network or
affect other members.`,
		Example: `  gocrypt network leave homelab`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return runNetworkLeave(cmd, global, args[0])
		},
	}
}

func runNetworkLeave(cmd *cobra.Command, global *cli.GlobalOptions, network string) error {
	return nil
}
