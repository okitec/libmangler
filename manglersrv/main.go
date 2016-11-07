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
	"time"
	"unicode"

	"github.com/okitec/libmangler/elem"
	"github.com/okitec/libmangler/sexps"
)

// Protocol constants
const (
	protoVersion   = 11
	protoPort      = 40000
	protoEndMarker = "---\n" // for determining end-of-response
)

const (
	autosaveTime = 10 * time.Minute // time between two autosaves
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

		case 'b':
			// input: A (Book 978-0-201-141-03614-4 (authors "George Orwell") (title "Nineteen Eighty-Four"))
			type bookinfo struct {
				isbn    string
				authors []string
				title   string

				isbnFilled bool
				state      string
				skip       int
			}
			bi := bookinfo{"", nil, "", false, "", 0}

			// Remove cmd A from line
			args := strings.Fields(s)
			t := strings.Join(args[1:], " ")
			sexp, _, err := sexps.Parse(t)
			if err != nil {
				return fmt.Sprintf("can't create book: parsing error: %v\n", err)
			}

			sexps.Apply(sexp, func(atom sexps.Sexp, parent sexps.Sexp, data interface{}) {
				b := data.(*bookinfo)

				if b.skip > 0 {
					b.skip--
					return
				}

				switch b.state {
				case "":
					switch atom.String() {
					case "book":
						b.state = "get-isbn"
					case "authors":
						b.authors = sexps.List(parent.Cdr())
						b.skip = len(b.authors)
					case "title":
						b.state = "get-title"
					default:
						// just ignore unknown atoms
					}

				case "get-isbn":
					b.isbn = atom.String()
					b.isbnFilled = true
					b.state = ""

				case "get-title":
					b.title = atom.String()
					b.state = ""
				}
			}, &bi)

			if !bi.isbnFilled {
				return fmt.Sprintf("can't create book: not even an ISBN? Really?\n")
			}

			_, err = elem.NewBook(bi.isbn, bi.title, bi.authors)
			if err != nil {
				return fmt.Sprintf("can't create book: NewBook: %v\n", err)
			}
			return ""

		case 'c':
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
				start++ // +1: skip the slash
				endslash := strings.IndexRune(s[start:], '/')
				if endslash < 0 {
					return "missing closing slash\n"
				}
				end = endslash + 2 // +2: used as slice end
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

			// Remove duplicates
			for i := 0; i < len(*dot); i++ {
				// Don't need to check indices 0 .. i because they were already tested.
				for j := i + 1; j < len(*dot); /* empty */ {
					if (*dot)[i] == (*dot)[j] {
						(*dot)[j] = (*dot)[len(*dot)-1]
						(*dot)[len(*dot)-1] = nil       // so that it garbage-collects
						*dot = (*dot)[:len(*dot)-1]
						j = i + 1
					} else {
						j++
					}
				}
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

	nbooks, nusers, ncopies, err := load()
	if err != nil {
		// If we couldn't open the file, it's nothing; create them on exit
		_, ok := err.(*os.PathError)
		if ok {
			log.Println("no input files found; new 'books', 'copies', and 'users' files will be created on close")
			log.Println("[ATTENZIONE, PREGO] Check whether the user manglersrv runs as has permission to create files in the current directory")
			log.Println("[ATTENZIONE, PREGO] and whether any existing 'books', 'copies', and 'users' files can be read and written.")
		} else {
			log.Printf("can't load data: %v", err)
			log.Fatalf("quitting without overwriting old data")
		}
	} else {
		log.Printf("loading data: %v books, %v users, %v copies", nbooks, nusers, ncopies)
	}

	//cf. https://golang.org/pkg/os/signal/#Notify
	//cf. http://stackoverflow.com/questions/11268943/golang-is-it-possible-to-capture-a-ctrlc-signal-and-run-a-cleanup-function-in
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	go func() {
		for _ = range sc {
			store("books", "copies", "users")
			log.Println("saved data by overwriting old data, exiting now...")
			os.Exit(0)
		}
	}()

	// Autosave loop
	tc := time.Tick(autosaveTime)
	go func(tc <-chan time.Time) {
		for _ = range tc {
			store("books.autosave", "copies.autosave", "user.autosave")
			log.Println("autosaving data to *.autosave")
		}
	}(tc)

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
