// Package cmd implements the gocrypt command line.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// globalOptions holds flags that every command accepts.
type globalOptions struct {
	home string // --home: the node's folder
}

// Execute runs the gocrypt command line. Cobra prints the error, so this only
// sets the exit status.
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

// newRootCmd builds the whole command tree. Building it fresh (instead of
// using package-level variables) gives every test its own flags and output.
func newRootCmd() *cobra.Command {
	global := &globalOptions{}

	root := &cobra.Command{
		Use:   "gocrypt",
		Short: "Encrypted distributed file storage for your home network",
		Long: `gocrypt is a distributed file-storage tool for your home network.

It splits files into chunks, encrypts them, and distributes
the encrypted data across participating storage nodes.

Use gocrypt to manage nodes, store files, inspect storage status,
and restore previously stored data.

Run "gocrypt --help" to see available commands, or
"gocrypt <command> --help" for details about a specific command.`,
	}

	// Persistent: every subcommand accepts --home.
	root.PersistentFlags().StringVar(&global.home, "home", "",
		"the node's folder (default: $GOCRYPT_HOME, or ~/.gocrypt)")

	root.AddCommand(newInitCmd(global))
	return root
}
