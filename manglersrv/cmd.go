package main

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// simpleCmdFn describes a 'simple' command that does not work on selections.
// Examples are A, a, u, v and q. Because we must not close the connection here
// (handle has the close deferred already), q is implemented in handle() directly.
type simpleCmdFn func(rw io.ReadWriter, args []string) error

// "It is the error implementation's responsibility to summarize the context."
// We don't do that here; the caller adds a "cmd %v" prefix for logging.

// XXX get rid of stupid "s := "..."; fmt.Fprintln(rw, s); return errors.New(s)" dances
var simpleCmdtab = map[rune]simpleCmdFn{
	'A': func(rw io.ReadWriter, args []string) error {
		if len(args) < 2 {
			s := "specify ISBN"
			fmt.Fprintln(rw, s)
			return errors.New(s)
		}

		_, err := NewBook(args[1], "foo", nil) // XXX fetch or ask for title and author
		return err
	},
	'a': func(rw io.ReadWriter, args []string) error {
		var b *Book
		var n int
		var err error
		var ok bool

		if len(args) < 3 {
			s := "specify ISBN and number of books"
			fmt.Fprintln(rw, s)
			return errors.New(s)
		}

		if b, ok = books[ISBN(args[1])]; !ok {
			s := "book doesn't exist"
			fmt.Fprintln(rw, s)
			return errors.New(s)
		}

		if n, err = strconv.Atoi(args[2]); err != nil {
			s := "not a number"
			fmt.Fprintln(rw, s)
			return errors.New(s)
		}

		for i := 0; i < n; i++ {
			NewCopy(b)
		}

		return nil
	},
	'u': func(rw io.ReadWriter, args []string) error {
		if len(args) < 2 {
			s := "specify username"
			fmt.Fprintln(rw, s)
			return errors.New(s)
		}

		// args[1:]: skip first element "u"
		name := strings.Join(args[1:], " ")
		_, err := NewUser(name)
		return err
	},
	'v': func(rw io.ReadWriter, args []string) error {
		fmt.Fprintf(rw, "libmangler proto %d\n", protoVersion)

		if len(args) < 2 {
			s := "specify your protocol version"
			fmt.Fprintln(rw, s)
			return errors.New(s)
		}

		i, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Fprintf(rw, "%q: not a number\n", args[1])
			return err
		}

		// XXX Tell rest of manglersrv that a mismatch is fatal. But how? Panic/recover?
		if i != protoVersion {
			s := fmt.Sprintf("version mismatch (server %d, client %d)", protoVersion, i)
			fmt.Fprintln(rw, s)
			return errors.New(s)
		}

		return nil
	},
}
