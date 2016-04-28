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

		s := string(buf[0:n])
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
				err := fn(rw, args)
				if err != nil {
					log.Printf("cmd %c: %v", r, err)
				}
				break parse

			case 'B', 'C', 'U':
				// B/.../, C/.../, U/.../ reset the selection.
				dot = nil

				// Split /foo, bar, quux,/ into ["foo" "bar" "quux"].
				start := strings.IndexRune(s, '/') + 1       // +1: skip the slash
				end := strings.IndexRune(s[start:], '/') + 2 // +2: used as slice end

				// XXX really hacky solution, doesn't fulfill spec (strings are single-quoted in spec)
				csvr := csv.NewReader(strings.NewReader(s[start:end]))

				args, err := csvr.Read()
				if err != nil {
					log.Println("bad selection argument")
					break parse
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

				// Skip the selection arg, i.e. everything between the slashes (/.../)
				s = s[end+1:]
				// Actually reset the range-loop, as we modify the string.
				// Else we'd have spurious looping (I tested it).
				goto parse

			case 'p':
				for _, e := range dot {
					fmt.Fprintln(rw, e.Print())
				}

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
	NewBook("978-0-201-07981-4", "The AWK Programming Language")
	users = make(map[string]*User)
	NewUser("Florian the Florist from Florida")
	NewUser("Gaius Valerius Catullus")
	NewUser("Drago Mafloy")
	copies = make(map[int64]*Copy)
	NewCopy(books["978-0-201-07981-4"])

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
