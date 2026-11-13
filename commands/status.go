package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/its-kos/gocrypt/internal/repo"
)

func newStatusCmd(global *GlobalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show this node's details and check it for problems",
		Long: `Show this node's details and check it for problems.

status never asks for a passphrase. If the node's key is encrypted, it is not
compared with the config here; that happens when the node starts.

The exit status is 1 if there is any problem, so status can be used in
scripts and health checks.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true
			return runStatus(cmd, global)
		},
	}
}

func runStatus(cmd *cobra.Command, global *GlobalOptions) error {
	home, err := repo.ResolveHome(global.home)
	if err != nil {
		return err
	}

	info, err := repo.Load(home)
	if errors.Is(err, repo.ErrNotInitialized) {
		return fmt.Errorf("no GoCrypt node in %s; set one up with: gocrypt init", home)
	}
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Node:     %s\n", info.Config.Node.Name)
	fmt.Fprintf(out, "Node ID:  %s\n", info.Config.Node.ID)
	fmt.Fprintf(out, "Folder:   %s\n", info.Home)
	fmt.Fprintf(out, "Storage:  %s\n", info.StoragePath)
	if encrypted, err := info.KeyEncrypted(); err == nil {
		if encrypted {
			fmt.Fprintln(out, "Key:      encrypted (not checked: that needs the passphrase)")
		} else {
			fmt.Fprintln(out, "Key:      not encrypted")
		}
	} // an unreadable key shows up in the problems below

	problems := info.Check()
	if len(problems) == 0 {
		fmt.Fprintln(out, "OK")
		return nil
	}

	// Print every problem, then fail, so scripts see a
	fmt.Fprintln(out, "Problems:")
	for _, p := range problems {
		fmt.Fprintf(out, "  - %v\n", p)
	}
	return fmt.Errorf("found %d problem(s) with the node", len(problems))
}
