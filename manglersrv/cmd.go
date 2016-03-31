package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
)

// simpleCmdFn describes a 'simple' command that does not work on selections.
// Examples are A, a, u, v and q. Because we must not close the connection here
// (handle has the close deferred already), q is implemented in handle() directly.
type simpleCmdFn func(conn net.Conn, args []string) error

// "It is the error implementation's responsibility to summarize the context."
// We don't do that here; the caller adds a "cmd %v" prefix for logging.

var simpleCmdtab = map[rune]simpleCmdFn{
	'A': func(conn net.Conn, args []string) error {
		return errors.New("unimplemented")
	},
	'a': func(conn net.Conn, args []string) error {
		return errors.New("unimplemented")
	},
	'u': func(conn net.Conn, args []string) error {
		return errors.New("unimplemented")
	},
	'v': func(conn net.Conn, args []string) error {
		fmt.Fprintf(conn, "libmangler proto %d\n", protoVersion)

		if len(args) < 2 {
			s := "specify your protocol version"
			fmt.Fprintln(conn, s)
			return errors.New(s)
		}

		i, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Fprintf(conn, "%q: not a number\n", args[1])
			return err
		}

		// XXX Tell rest of manglersrv that a mismatch is fatal. But how? Panic/recover?
		if i != protoVersion {
			s := fmt.Sprintf("version mismatch (server %d, client %d)", protoVersion, i)
			fmt.Fprintln(conn, s)
			return errors.New(s)
		}

		return nil
	},
}
