# GoCrypt

[![License: GPL v3](https://img.shields.io/badge/license-GPL--3.0-blue.svg)](LICENSE)
[![Go version](https://img.shields.io/github/go-mod/go-version/its-kos/gocrypt)](go.mod)

Encrypted file storage across the machines you own.

GoCrypt turns the spare disk space on your own computers (a desktop, a home
server, a laptop) into a single encrypted store. Files are encrypted before
they leave the machine and spread across your nodes, so you can get them back
from any of them.

> [!WARNING]
> GoCrypt is in early development. Node setup is being built now, and storing
> files is not possible yet. Do not use it for data you care about.

## Goals

- Encrypt everything before it leaves the machine.
- Store data only on machines you control. No cloud service, no central server.
- Restore files from any node, even if the machine that stored them is gone.
- One binary and one CLI on every machine.

## How it works

This describes the intended design. Most of it is not built yet.

- **Nodes.** Every machine that takes part runs GoCrypt and is called a node.
  There is no master node; any node can store and restore files.
- **CLI and node service.** The same `gocrypt` binary is used for short
  commands (`init`, `status`, storing and restoring files) and for a
  long-running node service that other nodes talk to. The service is always
  started explicitly, never in the background behind your back.
- **Identity.** On `gocrypt init`, each node generates an Ed25519 keypair. The
  public key, written in base32, is the node's ID. Nodes will use these keys to
  prove who they are to each other.
- **Encryption.** File data is encrypted with AES-256-GCM before it is sent to
  another node, so storage nodes only ever hold ciphertext.
- **Distribution.** Files are split into chunks and placed on different nodes.
  Redundancy, so a file survives losing a machine, is planned but not designed
  yet.

## Status

- [x] Node identity: key generation, node IDs, key file format with optional passphrase
- [x] Encryption, compression and chunking building blocks
- [ ] `gocrypt init`: local node setup (in progress)
- [ ] Pairing nodes into a network
- [ ] Node service
- [ ] Storing and restoring files
- [ ] Redundancy and repair

## Installation

Requires Go 1.27 or later. Linux and macOS are supported; Windows support is
planned.

```sh
go install github.com/its-kos/gocrypt/cmd/gocrypt@latest
```

Or build from source:

```sh
git clone https://github.com/its-kos/gocrypt.git
cd gocrypt
go build -o gocrypt ./cmd/gocrypt
```

## Usage

`gocrypt init` is being built. Once finished it will work like this:

```sh
gocrypt init                            # set up this machine as a node
gocrypt init --name homelab             # pick a node name (default: hostname)
gocrypt init --storage-path /mnt/data   # keep stored data on another disk
gocrypt init --encrypt-key              # protect the node key with a passphrase
```

Running `init` again on a machine that is already set up changes nothing and
prints the node's details.

By default a node lives in `~/.gocrypt`:

```
~/.gocrypt/
├── config.toml     node name, node ID, storage path
├── identity.key    the node's private key (readable only by you)
└── storage/        data this node stores
```

Use `--home` or the `GOCRYPT_HOME` environment variable to put it somewhere
else, for example to run several test nodes on one machine.

More commands will be documented here as they are added.

## Security

GoCrypt has not been audited, and its formats may still change. It uses
well-known algorithms from Go's standard library and `golang.org/x/crypto`
(Ed25519, AES-256-GCM, Argon2id) rather than custom cryptography.

If you find a security problem, please report it privately through
[GitHub security advisories](https://github.com/its-kos/gocrypt/security/advisories/new)
instead of opening a public issue.

## Development

```sh
go build ./...
go vet ./...
go test ./...
```

```
cmd/gocrypt/           main package
cmd/commands/          CLI commands (Cobra)
internal/identity/     node keypair, node IDs, key file format
internal/encryptor/    AES-256-GCM encryption
internal/compressor/   gzip compression
internal/chunker/      splitting data into chunks
internal/config/       configuration
```

## Contributing

GoCrypt is still taking shape and isn't ready for outside contributions.
Bug reports and ideas are welcome in the
[issue tracker](https://github.com/its-kos/gocrypt/issues).

## License

Copyright (C) 2026 its-kos

GoCrypt is free software, released under the GNU General Public License v3.0.
See [LICENSE](LICENSE) for the full text.
