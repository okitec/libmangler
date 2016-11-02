/*
Manglersrv does stuff.
*/
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"unicode"

	"github.com/okitec/libmangler/elem"
)

// Protocol constants
const (
	protoVersion   = 9
	protoPort      = 40000
	protoEndMarker = "---\n" // for determining end-of-response
)

// The function interpret executes the line and returns a string that should be
// sent to the client. This is also used by store(), which is why it was split off
// handle in the first place.
func interpret(s string, dot *[]elem.Elem) (ret string) {
	var err error

parse:
	// We actually modify s sometimes, which is why we need a parse goto-label.
	for _, r := range s {
		if unicode.IsSpace(r) {
			continue
		}

		switch r {
		case '\n':
			return ""

		case 'A':
			args := strings.Fields(s)
			if len(args) < 2 {
				return "Can't create book: missing ISBN\n"
			}

			_, err = elem.NewBook(args[1], "foo", nil) // XXX fetch or ask for title and author
			return ""

		case 'a':
			var b *elem.Book
			var n int
			var ok bool

			args := strings.Fields(s)
			if len(args) < 3 {
				return "Can't create copy: missing ISBN and number of books\n"
			}

			if b, ok = elem.Books[elem.ISBN(args[1])]; !ok {
				return "Can't create copy: book doesn't exist\n"
			}

			if n, err = strconv.Atoi(args[2]); err != nil {
				return "Can't create copy: count is not a number\n"
			}

			for i := 0; i < n; i++ {
				var id int64

				// Skip used ids
				for id = rand.Int63(); elem.Copies[id] != nil; id = rand.Int63() {
				}

				elem.NewCopy(id, b)
			}

			return ""

		case 'B', 'C', 'U':
			// B/.../, C/.../, U/.../ reset the selection.
			*dot = nil
			var args []string
			var end int

			args = nil
			hasarg := false

			// Is there a slash? If so, split /foo, bar, quux,/ into ["foo" "bar" "quux"].
			// If not, args will be nil and hasarg stays false.
			start := strings.IndexRune(s, '/')
			if start != -1 {
				hasarg = true
				start++                                     // +1: skip the slash
				end = strings.IndexRune(s[start:], '/') + 2 // +2: used as slice end
				csvr := csv.NewReader(strings.NewReader(s[start:end]))

				args, err = csvr.Read()
				if err != nil {
					log.Println("bad selection argument")
					return ""
				}
			}

			for i := range args {
				args[i] = strings.TrimSpace(args[i])
			}

			// Do not use := here, it would redefine dot. Subtle.
			*dot, err = elem.Select(r, *dot, args)
			if err != nil {
				log.Printf("cmd %c: %v", r, err)
			}

			// Skip the selection arg, i.e. everything between the slashes (/.../).
			if hasarg {
				s = s[end+1:]
				// Actually reset the range-loop, as we modify the string.
				// Else we'd have spurious looping (I tested it).
				goto parse
			}

		case 'd':
			for _, e := range *dot {
				e.Delete()
			}

			return ""

		case 'l':
			args := strings.Fields(s)
			name := strings.Join(args[1:], " ")
			name = strings.TrimRight(name, "\n")
			u := elem.Users[name]

			for _, e := range *dot {
				c, ok := e.(*elem.Copy)
				if !ok {
					log.Printf("tried to lend a non-Copy element")
					return "error: can't lend: not a Copy\n"
				}

				err = c.Lend(u)
				if err != nil {
					log.Printf("can't lend %v to user %v (%s)", c, u, name)
					return "error: can't lend: " + err.Error() + "\n"
				}
			}

			return ""

		case 'λ':
			sret := ""
			for _, e := range *dot {
				sret += e.List() + "\n"
			}

			return sret

		case 'n':
			// The note is all text after "n" and before the EOL, whitespace-trimmed.
			args := strings.Fields(s)
			note := strings.Join(args[1:], " ")
			note = strings.TrimRight(note, "\n")
			note = strings.TrimSpace(note)

			for _, e := range *dot {
				e.Note(note)
			}

			// No need to continue, the note is the rest of the line.
			return ""

		case 'p':
			sret := ""
			for _, e := range *dot {
				sret += e.Print() + "\n"
			}

			return sret

		case 'q':
			// We must handle QUIT here to avoid closing the connection both
			// in the deferred call and here. So send special quit string,
			// handle will know.
			return "quit"

		case 'R':
			for _, e := range *dot {
				e.Tag(true, "retired")
			}

		case 'r':
			for _, e := range *dot {
				c, ok := e.(*elem.Copy)
				if !ok {
					log.Printf("tried to return a non-Copy element")
					return "error: can't return: not a Copy\n"
				}

				c.Return()
			}

		case 'T':
			tags := make(map[string]int)

			for _, b := range elem.Books {
				for _, t := range b.Tags {
					tags[t]++
				}
			}

			for _, c := range elem.Copies {
				for _, t := range c.Tags {
					tags[t]++
				}
			}

			for _, u := range elem.Users {
				for _, t := range u.Tags {
					tags[t]++
				}
			}

			s := ""
			for t, _ := range tags {
				if t != "" {
					s += t + "\n"
				}
			}
			return s

		case 't':
			var add bool

			args := strings.Fields(s)
			if len(args) < 2 {
				log.Printf("missing tag arguments: %s", s)
				return "usage: t +|- tag\n"
			}

			if args[1] == "+" {
				add = true
			} else { // just assume it's a minus ('-')
				add = false
			}

			for _, e := range *dot {
				e.Tag(add, args[2])
			}

			return ""

		case 'u':
			args := strings.Fields(s)
			if len(args) < 2 {
				return "Can't create user: missing username\n"
			}

			// args[1:]: skip first element "u" (the command)
			name := strings.Join(args[1:], " ")
			_, err = elem.NewUser(name)
			if err != nil {
				return "Can't create user: " + err.Error() + "\n"
			}
			return ""

		case 'v':
			args := strings.Fields(s)
			sret := fmt.Sprintf("libmangler proto %d\n", protoVersion)

			if len(args) < 2 {
				return sret + "specify your protocol version\n"
			}

			i, err := strconv.Atoi(args[1])
			if err != nil {
				return sret + fmt.Sprintf("version is not a number (%q)\n", args[1])
			}

			// XXX Tell rest of manglersrv that a mismatch is fatal. But how? Panic/recover?
			if i != protoVersion {
				return sret + fmt.Sprintf("version mismatch (server %d, client %d)\n", protoVersion, i)
			}

			return sret
		}
	}

	return ""
}

// The function handle communicates with a client, resolving its requests via interpret().
func handle(rw io.ReadWriter) {
	var dot []elem.Elem
	var err error
	var n int
	buf := make([]byte, 128)

	for {
		n, err = rw.Read(buf)
		s := string(buf[0:n])
		ret := interpret(s, &dot)
		if ret == "quit" {
			return
		} else {
			fmt.Fprint(rw, ret)
			fmt.Fprint(rw, protoEndMarker)
		}
	}

	if err != nil {
		if err != io.EOF {
			log.Println("conn.Read error:", err)
		}

		return
	}
}

func main() {
	log.Println("libmangler proto", protoVersion)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", protoPort))
	if err != nil {
		log.Panicln("net.Listen failed:", err)
	}

	nbooks, nusers, ncopies := load()
	log.Printf("loading data: %v books, %v users, %v copies", nbooks, nusers, ncopies)

	//cf. https://golang.org/pkg/os/signal/#Notify
	//cf. http://stackoverflow.com/questions/11268943/golang-is-it-possible-to-capture-a-ctrlc-signal-and-run-a-cleanup-function-in
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	go func() {
		for _ = range sc {
			store()
			log.Println("Saved data, exiting now...")
			os.Exit(0)
		}
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("ln.Accept failed:", err)
			continue
		}

		log.Println("client", conn.RemoteAddr(), "connected")
		go func() {
			defer conn.Close()
			defer log.Println("client", conn.RemoteAddr(), "disconnected")
			handle(conn)
		}()
	}
}
