package network

import (
	cmd "github.com/its-kos/gocrypt/commands"
	"github.com/spf13/cobra"
)

func newNetworkCmd(global *cmd.globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "network",
		Short: "Show this node's details and check it for problems",
		Long: `Show this node's details and check it for problems.

status never asks for a passphrase. If the node's key is encrypted, it is not
compared with the config here; that happens when the node starts.

The exit status is 1 if there is any problem, so status can be used in
scripts and health checks.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true
			return runNetwork(cmd, global)
		},
	}
}

func newNetworkStatusCmd(global *cmd.globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "network",
		Short: "Show this node's details and check it for problems",
		Long: `Show this node's details and check it for problems.

status never asks for a passphrase. If the node's key is encrypted, it is not
compared with the config here; that happens when the node starts.

The exit status is 1 if there is any problem, so status can be used in
scripts and health checks.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true
			return runNetwork(cmd, global)
		},
	}
}

func newNetworkCreateCmd(global *cmd.globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "network",
		Short: "Show this node's details and check it for problems",
		Long: `Show this node's details and check it for problems.

status never asks for a passphrase. If the node's key is encrypted, it is not
compared with the config here; that happens when the node starts.

The exit status is 1 if there is any problem, so status can be used in
scripts and health checks.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true
			return runNetwork(cmd, global)
		},
	}
}

func runNetwork(cmd *cobra.Command, global *cmd.globalOptions) error {}
