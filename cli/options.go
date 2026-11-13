// Package cli handles shared CLI types to avoid cyclical imports
package cli

// GlobalOptions holds flags that every command accepts.
type GlobalOptions struct {
	Home string // --home: the node's folder. Used to run each command in that specific home.
}
