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
)

// Protocol constants
const (
	protoVersion = -1
	protoPort    = 40000
)

// The function interpret executes the line and returns a string that should be
// sent to the client. This is also used by store(), which is why it was split off
// handle in the first place.
func interpret(s string, dot *[]elem) (ret string) {
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

			fn := seltab[r]
			// Do not use := here, it would redefine dot. Subtle.
			*dot, err = fn(*dot, args)
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
			u := users[name]

			for _, e := range *dot {
				c, ok := e.(*Copy)
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
				c, ok := e.(*Copy)
				if !ok {
					log.Printf("tried to return a non-Copy element")
					return "error: can't return: not a Copy\n"
					break
				}

				c.Return()
			}
		}
	}

	return ""
}

// The function handle communicates with a client, resolving its requests via interpret().
func handle(rw io.ReadWriter) {
	var dot []elem
	var err error
	var n int
	buf := make([]byte, 128)

	for {
		n, err = rw.Read(buf)
		s := string(buf[0:n])
		log.Printf("[->proto] %q\n", s)
		ret := interpret(s, &dot)
		if ret == "quit" {
			return
		} else {
			fmt.Fprint(rw, ret)
			log.Printf("[proto->] %q\n", ret)
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

	books = make(map[ISBN]*Book)
	users = make(map[string]*User)
	copies = make(map[int64]*Copy)

	// debugging examples
	NewBook("978-0-201-07981-4", "The AWK Programming Language", []string{"Alfred V. Aho", "Brian W. Kernighan", "Peter J. Weinberger"})
	NewBook("978-0-141-03614-4", "Nineteen Eighty-Four", []string{"George Orwell"})
	u1, _ := NewUser("Florian the Florist from Florida")
	u2, _ := NewUser("Gaius Valerius Catullus")
	u3, _ := NewUser("Drago Mafloy")
	c1, _ := NewCopy(books["978-0-201-07981-4"])
	c2, _ := NewCopy(books["978-0-201-07981-4"])
	c3, _ := NewCopy(books["978-0-141-03614-4"])
	c4, _ := NewCopy(books["978-0-141-03614-4"])
	c1.Lend(u1)
	c2.Lend(u2)
	c3.Lend(u2)
	c4.Lend(u3)

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
