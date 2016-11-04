package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/okitec/libmangler/elem"
	"github.com/okitec/libmangler/sexps"
)

type copyctx struct {
	state         string
	idFilled      bool
	userFilled    bool
	isbnFilled    bool
	authorsFilled bool
	titleFilled   bool
	notesFilled   bool
	tagsFilled    bool
	skip          int // # of atoms to be skipped after sexps.List()

	id      int64
	user    string
	isbn    string
	authors []string
	title   string
	notes   []string
	tags    []string
}

type userctx struct {
	state        string
	nameFilled   bool
	notesFilled  bool
	tagsFilled   bool
	copiesFilled bool
	skip         int // # of atoms to be skipped after sexps.List()

	name   string
	notes  []string
	tags   []string
	copies []int64
}

type bookctx struct {
	state         string
	isbnFilled    bool
	authorsFilled bool
	titleFilled   bool
	notesFilled   bool
	tagsFilled    bool
	copiesFilled  bool
	skip          int // # of atoms to be skipped after sexps.List()

	isbn    string
	authors []string
	title   string
	notes   []string
	tags    []string
	copies  []int64
}

func handleCopy(atom sexps.Sexp, parent sexps.Sexp, data interface{}) {
	c := data.(*copyctx)

	if c.skip > 0 {
		c.skip--
		return
	}

	switch c.state {
	case "":
		if c.idFilled && c.userFilled && c.isbnFilled && c.authorsFilled && c.titleFilled && c.notesFilled && c.tagsFilled {
			c.state = "end"
			break
		}

		switch atom.String() {
		case "copy":
			c.state = "get-id"
		case "user":
			c.state = "get-user"
		case "book":
			c.state = "get-isbn"
		case "authors":
			c.authors = sexps.List(parent.Cdr())
			c.skip = len(c.authors)
			c.authorsFilled = true
		case "title":
			c.state = "get-title"
		case "notes":
			c.notes = sexps.List(parent.Cdr())
			c.skip = len(c.notes)
			c.notesFilled = true
		case "tags":
			c.tags = sexps.List(parent.Cdr())
			c.skip = len(c.tags)
			c.tagsFilled = true
		default:
			c.state = "err"
		}

	case "get-id":
		var err error
		c.id, err = strconv.ParseInt(atom.String(), 10, 64)
		if err != nil {
			c.state = "err"
		}

		c.idFilled = true
		c.state = ""

	case "get-user":
		c.user = atom.String()
		c.userFilled = true
		c.state = ""

	case "get-isbn":
		c.isbn = atom.String()
		c.isbnFilled = true
		c.state = ""

	case "get-title":
		c.title = atom.String()
		c.titleFilled = true
		c.state = ""

	case "end":
		return

	case "err":
	default:
		fmt.Println("bad state: " + c.state)
	}
}

func handleUser(atom sexps.Sexp, parent sexps.Sexp, data interface{}) {
	u := data.(*userctx)

	if u.skip > 0 {
		u.skip--
		return
	}

	switch u.state {
	case "":
		if u.nameFilled && u.notesFilled && u.tagsFilled && u.copiesFilled {
			u.state = "end"
			break
		}

		switch atom.String() {
		case "user":
			u.state = "get-name"
		case "notes":
			u.notes = sexps.List(parent.Cdr())
			u.skip = len(u.notes)
			u.notesFilled = true
		case "copies":
			var err error
			u.copies, err = getCopies(parent.Cdr())
			if err != nil {
				u.state = "err"
			}
			u.skip = len(u.copies)
			u.copiesFilled = true
		case "tags":
			u.tags = sexps.List(parent.Cdr())
			u.skip = len(u.tags)
			u.tagsFilled = true
		default:
			u.state = "err"
		}

	case "get-name":
		u.name = atom.String()
		u.nameFilled = true
		u.state = ""

	case "end":
		return

	case "err":
	default:
		fmt.Println("bad state: " + u.state)
	}
}

func handleBook(atom sexps.Sexp, parent sexps.Sexp, data interface{}) {
	b := data.(*bookctx)

	if b.skip > 0 {
		b.skip--
		return
	}

	switch b.state {
	case "":
		if b.isbnFilled && b.authorsFilled && b.titleFilled && b.notesFilled && b.tagsFilled && b.copiesFilled {
			b.state = "end"
			break
		}

		switch atom.String() {
		case "book":
			b.state = "get-isbn"
		case "authors":
			b.authors = sexps.List(parent.Cdr())
			b.skip = len(b.authors)
			b.authorsFilled = true
		case "title":
			b.state = "get-title"
		case "notes":
			b.notes = sexps.List(parent.Cdr())
			b.skip = len(b.notes)
			b.notesFilled = true
		case "tags":
			b.tags = sexps.List(parent.Cdr())
			b.skip = len(b.tags)
			b.tagsFilled = true
		case "copies":
			var err error
			b.copies, err = getCopies(parent.Cdr())
			if err != nil {
				b.state = "err"
			}
			b.skip = len(b.copies)
			b.copiesFilled = true
		default:
			b.state = "err"
		}

	case "get-isbn":
		b.isbn = atom.String()
		b.isbnFilled = true
		b.state = ""

	case "get-title":
		b.title = atom.String()
		b.titleFilled = true
		b.state = ""

	case "end":
		return

	case "err":
	default:
		fmt.Println("bad state: " + b.state)
	}
}

// getCopies puts an s-expr (copies 405 4959 ...) into []int64{405, 4959, ...}.
func getCopies(sexp sexps.Sexp) (ls []int64, err error) {
	scopies := sexps.List(sexp)
	for _, s := range scopies {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return ls, err
		}
		ls = append(ls, i)
	}

	return ls, nil
}

func load() (nbooks, nusers, ncopies int, err error) {
	// Copies must be read last so that they can be connected to their users.
	fnames := []string{"books", "users", "copies"}

	for _, fname := range fnames {
		input, err := os.Open(fname)
		if err != nil {
			return 0, 0, 0, err
		}

		// slurp whole file
		buf, err := ioutil.ReadAll(input)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("io error when reading '%s': %v", fname, err)
		}
		input.Close()

		tail := string(buf)
		for len(tail) > 1 { // there's a lonely newline never parsed
			var sexp sexps.Sexp
			var err error
			sexp, tail, err = sexps.Parse(tail)
			if err != nil {
				pos := len(string(buf)) - len(tail)
				nlines := strings.Count(string(buf[:pos]), "\n") + 1
				return nbooks, nusers, ncopies, fmt.Errorf("parse error in '%s', line %d: %v", fname, nlines, err)
			}

			switch fname {
			case "books":
				b := bookctx{}
				sexps.Apply(sexp, handleBook, &b)
				if b.state == "err" {
					pos := len(string(buf)) - len(tail)
					nlines := strings.Count(string(buf[:pos]), "\n") + 1
					return nbooks, nusers, ncopies, fmt.Errorf("malformed book entry around line %d in '%s'", nlines, fname)
				}

				bp, err := elem.NewBook(b.isbn, b.title, b.authors)
				if err != nil {
					return nbooks, nusers, ncopies, fmt.Errorf("can't create book: %v", err)
				}

				// Must be last, clears any notes produced by NewBook.
				bp.Notes = b.notes
				bp.Tags = b.tags

				nbooks++

			case "users":
				u := userctx{}
				sexps.Apply(sexp, handleUser, &u)
				if u.state == "err" {
					pos := len(string(buf)) - len(tail)
					nlines := strings.Count(string(buf[:pos]), "\n") + 1
					return nbooks, nusers, ncopies, fmt.Errorf("malformed user entry around line %d in '%s'", nlines, fname)
				}

				up, err := elem.NewUser(u.name)
				if err != nil {
					return nbooks, nusers, ncopies, fmt.Errorf("can't create user: %v", err)
				}

				// Must be last, clears any notes produced by NewUser.
				up.Notes = u.notes
				up.Tags = u.tags

				nusers++

			case "copies":
				c := copyctx{}
				sexps.Apply(sexp, handleCopy, &c)
				if c.state == "err" {
					pos := len(string(buf)) - len(tail)
					nlines := strings.Count(string(buf[:pos]), "\n") + 1
					return nbooks, nusers, ncopies, fmt.Errorf("malformed copy entry around line %d in '%s'", nlines, fname)
				}

				cp, err := elem.NewCopy(c.id, elem.Books[elem.ISBN(c.isbn)])
				if err != nil {
					return nbooks, nusers, ncopies, fmt.Errorf("can't create copy: %v", err)
				}

				if len(c.user) > 0 && elem.Users[c.user] != nil {
					err := cp.Lend(elem.Users[c.user])
					if err != nil {
						return nbooks, nusers, ncopies, fmt.Errorf("can't associate copy with user anymore: %v", err)
					}
				}

				// Must be last, clears any notes produced by NewCopy and Lend.
				cp.Notes = c.notes
				cp.Tags = c.tags

				ncopies++
			}
		}
	}

	return nbooks, nusers, ncopies, nil
}
