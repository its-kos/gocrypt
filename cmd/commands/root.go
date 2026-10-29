package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
