package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

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

	isbn    string
	authors []string
	title   string
	notes   []string
	tags    []string
	copies  []int64
}

func handleCopy(atom sexps.Sexp, parent sexps.Sexp, data interface{}) {
	c := data.(*copyctx)

	switch c.state {
	case "":
		if c.idFilled && c.userFilled && c.isbnFilled && c.authorsFilled && c.titleFilled && c.notesFilled && c.tagsFilled {
			c.state = "end"
			break
		}

		switch atom.String() {
		case "copy":
			c.state = "copy"
		case "user":
			c.state = "user"
		case "book":
			c.state = "book"
		case "notes":
			c.state = "notes"
		case "tags":
			c.state = "tags"
		default:
			c.state = "err"
		}

	case "copy":
		var err error
		c.id, err = strconv.ParseInt(atom.String(), 10, 64)
		if err != nil {
			c.state = "err"
		}

		c.idFilled = true
		c.state = ""

	case "user":
		c.user = atom.String()
		c.userFilled = true
		c.state = ""

	case "book":
		c.isbn = atom.String()
		c.isbnFilled = true
		c.state = "book2"

	case "book2":
		if c.authorsFilled && c.titleFilled {
			c.state = ""
			break
		}

		switch atom.String() {
		case "authors":
			c.state = "authors"
		case "title":
			c.state = "title"
		default:
			// do nothing; the authors match here
		}

	case "authors":
		c.authors = sexps.List(parent)
		c.authorsFilled = true
		if !c.titleFilled {
			c.state = "book2"
		} else {
			c.state = ""
		}

	case "title":
		c.title = atom.String()
		c.titleFilled = true
		if !c.authorsFilled {
			c.state = "book2"
		} else {
			c.state = ""
		}

	case "notes":
		c.notes = sexps.List(parent)
		c.notesFilled = true
		c.state = ""

	case "tags":
		c.tags = sexps.List(parent)
		c.tagsFilled = true
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

	switch u.state {
	case "":
		if u.nameFilled && u.notesFilled && u.tagsFilled && u.copiesFilled {
			u.state = "end"
			break
		}

		switch atom.String() {
		case "user":
			u.state = "user"
		case "notes":
			u.state = "notes"
		case "copies":
			u.state = "copies"
		case "tags":
			u.state = "tags"
		default:
			u.state = "err"
		}

	case "user":
		u.name = atom.String()
		u.nameFilled = true
		u.state = ""

	case "notes":
		// In this state, we are at the atom "notes". sexps.List(parent)
		// would include "notes" as an item of the list.
		u.notes = sexps.List(parent.Cdr())
		u.notesFilled = true
		u.state = ""

	// XXX doesn't work - maybe because we never enter this state. Test!
	case "tags":
		u.tags = sexps.List(parent)
		u.tagsFilled = true
		u.state = ""

	case "copies":
		var err error
		u.copies, err = getCopies(parent)
		if err != nil {
			u.state = "err"
		} else {
			u.copiesFilled = true
			u.state = ""
		}

	case "end":
		return

	case "err":
	default:
		fmt.Println("bad state: " + u.state)
	}
}

// XXX heavily duplicated from handleCopy
func handleBook(atom sexps.Sexp, parent sexps.Sexp, data interface{}) {
	b := data.(*bookctx)

	switch b.state {
	case "":
		if b.isbnFilled && b.authorsFilled && b.titleFilled && b.notesFilled && b.tagsFilled && b.copiesFilled {
			b.state = "end"
			break
		}

		switch atom.String() {
		case "book":
			b.state = "book"
		case "notes":
			b.state = "notes"
		case "tags":
			b.state = "tags"
		case "copies":
			b.state = "copies"
		default:
			b.state = "err"
		}

	case "book":
		b.isbn = atom.String()
		b.isbnFilled = true
		b.state = "book2"

	case "book2":
		if b.authorsFilled && b.titleFilled {
			b.state = ""
			break
		}

		switch atom.String() {
		case "authors":
			b.state = "authors"
		case "title":
			b.state = "title"
		default:
			// do nothing; the authors match here
		}

	case "authors":
		b.authors = sexps.List(parent)
		b.authorsFilled = true
		if !b.titleFilled {
			b.state = "book2"
		} else {
			b.state = ""
		}

	case "title":
		b.title = atom.String()
		b.titleFilled = true
		if !b.authorsFilled {
			b.state = "book2"
		} else {
			b.state = ""
		}

	case "notes":
		b.notes = sexps.List(parent)
		b.notesFilled = true
		b.state = ""

	case "tags":
		b.tags = sexps.List(parent)
		b.tagsFilled = true
		b.state = ""

	case "copies":
		var err error
		b.copies, err = getCopies(parent)
		if err != nil {
			b.state = "err"
		} else {
			b.copiesFilled = true
			b.state = ""
		}

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

func load() (nbooks, nusers, ncopies int) {
	// Copies must be read last so that they can be connected to their users.
	fnames := []string{"books", "users", "copies"}

	for _, fname := range fnames {
		input, err := os.Open(fname)
		if err != nil {
			log.Printf("can't open %s: %v", input, err)
			return
		}

		// slurp whole file
		buf, err := ioutil.ReadAll(input)
		if err != nil {
			log.Printf("io error when reading %s: %v", input, err)
			return
		}
		input.Close()

		tail := string(buf)
		for len(tail) > 1 { // there's a lonely newline never parsed
			var sexp sexps.Sexp
			var err error
			sexp, tail, err = sexps.Parse(tail)
			if err != nil {
				log.Printf("parse error in %s: %v", input, err)
				return
			}

			switch fname {
			case "books":
				b := bookctx{}
				sexps.Apply(sexp, handleBook, &b)
				bp, err := elem.NewBook(b.isbn, b.title, b.authors)
				if err != nil {
					log.Printf("can't create book: %v", err)
					break
				}

				// Must be last, clears any notes produced by NewBook.
				bp.Notes = b.notes
				bp.Tags = b.tags

				nbooks++

			case "users":
				u := userctx{}
				sexps.Apply(sexp, handleUser, &u)
				up, err := elem.NewUser(u.name)
				if err != nil {
					log.Printf("can't create user: %v", err)
					break
				}

				// Must be last, clears any notes produced by NewUser.
				up.Notes = u.notes
				up.Tags = u.tags

				nusers++

			case "copies":
				c := copyctx{}
				sexps.Apply(sexp, handleCopy, &c)
				cp, err := elem.NewCopy(c.id, elem.Books[elem.ISBN(c.isbn)])
				if err != nil {
					log.Printf("can't create copy: %v", err)
					break
				}

				if len(c.user) > 0 && elem.Users[c.user] != nil {
					err := cp.Lend(elem.Users[c.user])
					if err != nil {
						log.Printf("can't associate copy with user anymore: %v", err)
						break
					}
				}

				// Must be last, clears any notes produced by NewCopy and Lend.
				cp.Notes = c.notes
				cp.Tags = c.tags

				ncopies++
			}
		}
	}

	return
}
