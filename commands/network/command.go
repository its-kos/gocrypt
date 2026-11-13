// Package network provides commands for managing GoCrypt networks.
package network

import (
	"github.com/its-kos/gocrypt/cli"
	"github.com/spf13/cobra"
)

func NewNetworkCmd(global *cli.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network",
		Short: "Manage GoCrypt networks",
		Long: `Manage the logical networks this node belongs to.

Create networks, join existing networks, inspect network details,
and manage this node's network memberships.

Use "gocrypt network --help" to see available operations.`,
		Args: cobra.NoArgs,
	}

	cmd.AddCommand(newNetworkListCmd(global))
	cmd.AddCommand(newNetworkCreateCmd(global))
	cmd.AddCommand(newNetworkJoinCmd(global))
	cmd.AddCommand(newNetworkLeaveCmd(global))
	cmd.AddCommand(newNetworkInspectCmd(global))

	return cmd
}
