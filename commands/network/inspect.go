package network

import (
	"github.com/its-kos/gocrypt/cli"
	"github.com/spf13/cobra"
)

func newNetworkInspectCmd(global *cli.GlobalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <network>",
		Short: "Show details about a network",
		Long: `Show detailed information about a network.

Displays information such as the network ID, name, and known
membership details.`,
		Example: `  gocrypt network inspect homelab`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return runNetworkInspect(cmd, global, args[0])
		},
	}
}

func runNetworkInspect(cmd *cobra.Command, global *cli.GlobalOptions, network string) error {
	return nil
}
