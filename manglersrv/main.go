/*
Manglersrv does stuff.
*/
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"unicode"

	"github.com/okitec/libmangler/elems"
)

// Protocol constants
const (
	protoVersion   = 7
	protoPort      = 40000
	protoEndMarker = "---\n" // for determining end-of-response
)

// The function interpret executes the line and returns a string that should be
// sent to the client. This is also used by store(), which is why it was split off
// handle in the first place.
func interpret(s string, dot *[]elems.Elem) (ret string) {
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

		case 'q':
			// We must handle QUIT here to avoid closing the connection both
			// in the deferred call and in cmd.go. So send special quit string,
			// handle will know.
			return "quit"

		case 'A', 'a', 'u', 'v':
			// 'Simple' commands not operating on selections.
			args := strings.Fields(s)

			fn := simpleCmdtab[r]
			sret, err := fn(args)
			if err != nil {
				log.Printf("cmd %c: %v", r, err)
			}

			return sret + "\n"
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
				// XXX really hacky solution, doesn't fulfill spec (strings are single-quoted in spec)
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
			*dot, err = elems.Select(r, *dot, args)
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

		case 'p':
			sret := ""
			for _, e := range *dot {
				sret += e.Print() + "\n"
			}

			return sret
		case 'n':
			// The note is all text after "n" and before the EOL, whitespace-trimmed.
			note := s[1:strings.IndexRune(s, '\n')]
			note = strings.TrimSpace(note)
			for _, e := range *dot {
				e.Note(note)
			}

			// No need to continue, the note is the rest of the line.
			return ""

		case 'd':
			for _, e := range *dot {
				e.Delete()
			}

			return ""

		case 'l':
			name := s[1:strings.IndexRune(s, '\n')]
			name = strings.TrimSpace(name)
			u := elems.Users[name]

			for _, e := range *dot {
				c, ok := e.(*elems.Copy)
				if !ok {
					log.Printf("tried to lend a non-Copy element")
					return "error: can't lend: not a Copy\n"
					break
				}

				c.Lend(u)
			}

			return ""

		case 'r':
			for _, e := range *dot {
				c, ok := e.(*elems.Copy)
				if !ok {
					log.Printf("tried to return a non-Copy element")
					return "error: can't return: not a Copy\n"
				}

				c.Return()
			}

		case 't':
			var add bool

			args := strings.Fields(s)
			if len(args) < 2 {
				log.Printf("")
				return "usage: t +|- tag"
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

		case 'T':
			tags := make(map[string]int)

			for _, b := range elems.Books {
				for _, t := range b.Tags {
					tags[t]++
				}
			}

			for _, c := range elems.Copies {
				for _, t := range c.Tags {
					tags[t]++
				}
			}

			for _, u := range elems.Users {
				for _, t := range u.Tags {
					tags[t]++
				}
			}

			s := ""
			for t, _ := range tags {
				s += t + "\n"
			}
			return s
		}
	}

	return ""
}

// The function handle communicates with a client, resolving its requests via interpret().
func handle(rw io.ReadWriter) {
	var dot []elems.Elem
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

	elems.Books = make(map[elems.ISBN]*elems.Book)
	elems.Users = make(map[string]*elems.User)
	elems.Copies = make(map[int64]*elems.Copy)

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
