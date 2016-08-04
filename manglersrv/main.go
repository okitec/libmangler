/*
Manglersrv implements the v, q, B, p, n, and d commands at the moment.
Only Books are implemented; only ISBNs can be selected. Copies and Users
don't yet exist. The specification is violated by the CSV Reader for command B.
*/
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"unicode"
)

// Protocol constants
const (
	protoVersion = -1
	protoPort    = 40000
)

// The function handle communicates with a client, resolving its requests.
func handle(rw io.ReadWriter) {
	var dot []elem
	buf := make([]byte, 128)

	for {
		n, err := rw.Read(buf)

		req := string(buf[0:n])
		args := strings.Split(req, " ")
		if len(args) < 2 {
			log.Printf("request without tag: %s", req)
			fmt.Fprintf(rw, "error: request without tag: %s\n", req) // XXX breaks Java side
			// XXX duplication, see end of handle()
			if err != nil {
				return
			}
			continue
		}

		tag, err := strconv.Atoi(args[0])
		if err != nil {
			log.Printf("non-numerical tag %s", args[0])
			fmt.Fprintf(rw, "error: non-numerical tag %s\n", args[0]) // XXX breaks Java side
			continue
		}

		// stitch rest of request back together, excluding tag
		s := strings.Join(args[1:len(args)], " ")

	parse:
		// We actually modify s sometimes.
		for _, r := range s {
			if unicode.IsSpace(r) {
				continue
			}

			switch r {
			case '\n':
				break parse

			case 'q':
				// We must handle QUIT here to avoid closing the connection both
				// in the deferred call and in cmd.go.
				return

			case 'A', 'a', 'u', 'v':
				// 'Simple' commands not operating on selections.
				args := strings.Fields(s)

				fn := simpleCmdtab[r]
				sret, err := fn(args)
				if err != nil {
					log.Printf("cmd %c: %v", r, err)
				}

				sret += "\n"
				fmt.Fprintf(rw, "%d %d\n%s", tag, strings.Count(sret, "\n"), sret)
				break parse

			case 'B', 'C', 'U':
				// B/.../, C/.../, U/.../ reset the selection.
				dot = nil
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
						break parse
					}
				}

				for i := range args {
					args[i] = strings.TrimSpace(args[i])
				}

				fn := seltab[r]
				// Do not use := here, it would redefine dot. Subtle.
				dot, err = fn(dot, args)
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
				for _, e := range dot {
					sret += e.Print() + "\n"
				}
				fmt.Fprintf(rw, "%d %d\n", tag, strings.Count(sret, "\n"))
				fmt.Fprint(rw, sret)

			case 'n':
				// The note is all text after "n" and before the EOL, whitespace-trimmed.
				note := s[1:strings.IndexRune(s, '\n')]
				note = strings.TrimSpace(note)
				for _, e := range dot {
					e.Note(note)
				}

				// No need to continue, the note is the rest of the line.
				break parse

			case 'd':
				for _, e := range dot {
					e.Delete()
				}

			case 'l':
				name := s[1:strings.IndexRune(s, '\n')]
				name = strings.TrimSpace(name)
				u := users[name]

				for _, e := range dot {
					c, ok := e.(*Copy)
					if !ok {
						log.Printf("tried to lend a non-Copy element")
						fmt.Fprintf(rw, "%d 1\nerror: can't lend: not a Copy\n", tag)
						break
					}

					c.Lend(u)
				}

				break parse

			case 'r':
				for _, e := range dot {
					c, ok := e.(*Copy)
					if !ok {
						log.Printf("tried to return a non-Copy element")
						fmt.Fprintf(rw, "%d 1\nerror: can't return: not a Copy\n", tag)
						break
					}

					c.Return()
				}
			}
		}

		if err != nil {
			if err != io.EOF {
				log.Println("conn.Read error:", err)
			}

			return
		}
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
