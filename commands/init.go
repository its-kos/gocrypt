package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/its-kos/gocrypt/cli"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/its-kos/gocrypt/internal/repo"
)

// passphraseFileEnv names a file holding the passphrase, for when no one is
// there to type it (scripts, Docker secrets).
const passphraseFileEnv = "GOCRYPT_PASSPHRASE_FILE"

type initOptions struct {
	name           string
	storagePath    string
	encryptKey     bool
	passphraseFile string
}

func newInitCmd(global *cli.GlobalOptions) *cobra.Command {
	opts := &initOptions{}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Set up this machine as a GoCrypt node",
		Long: `Set up this machine as a GoCrypt node.

init creates the node's folder with a new identity key, a config file and a
storage folder. It only prepares this machine; it doesn't join a network.

Running init on a machine that is already set up changes nothing and prints
the node's details.`,
		Example: `  gocrypt init
  gocrypt init --name homelab --storage-path /mnt/data/gocrypt
  gocrypt init --home /tmp/test-node
  gocrypt init --encrypt-key --passphrase-file /run/secrets/gocrypt`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// From here on, errors are about the node, not about how the
			// command was typed, so don't print the usage text with them.
			cmd.SilenceUsage = true
			return runInit(cmd, global, opts)
		},
	}

	f := cmd.Flags()
	f.StringVar(&opts.name, "name", "", "a name for this node (default: the hostname)")
	f.StringVar(&opts.storagePath, "storage-path", "", "where this node keeps stored data (default: storage/ in the node's folder)")
	f.BoolVar(&opts.encryptKey, "encrypt-key", false, "protect the node's key with a passphrase (needed again every time the node starts)")
	f.StringVar(&opts.passphraseFile, "passphrase-file", "", "read the passphrase from this file (default: $"+passphraseFileEnv+")")
	return cmd
}

func runInit(cmd *cobra.Command, global *cli.GlobalOptions, opts *initOptions) error {
	if opts.passphraseFile != "" && !opts.encryptKey {
		return errors.New("--passphrase-file only makes sense together with --encrypt-key")
	}

	home, err := repo.ResolveHome(global.Home)
	if err != nil {
		return err
	}

	initOpts := repo.InitOptions{Name: opts.name, StoragePath: opts.storagePath}
	if opts.encryptKey {
		// Called by repo only if a node is really created.
		initOpts.Passphrase = func() ([]byte, error) {
			return readPassphrase(cmd, opts.passphraseFile)
		}
	}

	info, created, err := repo.Init(home, initOpts)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	if created {
		fmt.Fprintf(out, "Initialized GoCrypt node in %s\n", info.Home)
	} else {
		fmt.Fprintf(out, "GoCrypt node already initialized in %s (nothing changed)\n", info.Home)
	}
	fmt.Fprintf(out, "  Name:     %s\n", info.Config.Node.Name)
	fmt.Fprintf(out, "  Node ID:  %s\n", info.Config.Node.ID)
	fmt.Fprintf(out, "  Storage:  %s\n", info.StoragePath)
	return nil
}

// readPassphrase gets the passphrase from --passphrase-file, then
// $GOCRYPT_PASSPHRASE_FILE, then by asking twice on the terminal. The
// passphrase is never accepted as a flag value: it would end up in shell
// history and be visible to other users in the process list.
func readPassphrase(cmd *cobra.Command, file string) ([]byte, error) {
	if file == "" {
		file = os.Getenv(passphraseFileEnv)
	}
	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		// Files written with echo or an editor end in a newline that isn't
		// part of the passphrase.
		return []byte(strings.TrimRight(string(data), "\r\n")), nil
	}

	// Never wait for input that can't come (scripts, Docker, CI).
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return nil, fmt.Errorf("--encrypt-key needs a passphrase: run in a terminal, or use --passphrase-file or $%s", passphraseFileEnv)
	}

	prompt := cmd.ErrOrStderr()
	fmt.Fprint(prompt, "Passphrase for the node key: ")
	first, err := term.ReadPassword(fd)
	fmt.Fprintln(prompt)
	if err != nil {
		return nil, err
	}
	fmt.Fprint(prompt, "Repeat the passphrase: ")
	second, err := term.ReadPassword(fd)
	fmt.Fprintln(prompt)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(first, second) {
		return nil, errors.New("the passphrases don't match")
	}
	return first, nil
}
