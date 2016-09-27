package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// simpleCmdFn describes a 'simple' command that does not work on selections.
// Examples are A, a, u, v and q. Because we must not close the connection here
// (handle has the close deferred already), q is implemented in handle() directly.
// The returned string is sent over the line to the client.
type simpleCmdFn func(args []string) (s string, err error)

// "It is the error implementation's responsibility to summarize the context."
// We don't do that here; the caller adds a "cmd %v" prefix for logging.

// XXX get rid of stupid "s := "..."; return s, errors.New(s)" dances
var simpleCmdtab = map[rune]simpleCmdFn{
	'A': func(args []string) (s string, err error) {
		if len(args) < 2 {
			s = "specify ISBN"
			err = errors.New(s)
			return s, err
		}

		_, err = NewBook(args[1], "foo", nil) // XXX fetch or ask for title and author
		return "", err
	},
	'a': func(args []string) (s string, err error) {
		var b *Book
		var n int
		var ok bool

		if len(args) < 3 {
			s = "specify ISBN and number of books"
			err = errors.New(s)
			return s, err
		}

		if b, ok = books[ISBN(args[1])]; !ok {
			s = "book doesn't exist"
			err = errors.New(s)
			return s, err
		}

		if n, err = strconv.Atoi(args[2]); err != nil {
			s = "not a number"
			return s, err
		}

		for i := 0; i < n; i++ {
			NewCopy(b)
		}

		return "", err
	},
	'u': func(args []string) (s string, err error) {
		if len(args) < 2 {
			s := "specify username"
			err = errors.New(s)
			return s, err
		}

		// args[1:]: skip first element "u"
		name := strings.Join(args[1:], " ")
		_, err = NewUser(name)
		return "", err
	},
	'v': func(args []string) (s string, err error) {
		s = fmt.Sprintf("libmangler proto %d", protoVersion)

		if len(args) < 2 {
			s += "\nspecify your protocol version"
			err = errors.New("proto version")
			return s, err
		}

		i, err := strconv.Atoi(args[1])
		if err != nil {
			s += fmt.Sprintf("\n%q: not a number", args[1])
			return s, err
		}

		// XXX Tell rest of manglersrv that a mismatch is fatal. But how? Panic/recover?
		if i != protoVersion {
			s += fmt.Sprintf("\nversion mismatch (server %d, client %d)", protoVersion, i)
			err = errors.New("version mismatch")
			return s, err
		}

		return s, nil
	},
}
