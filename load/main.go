package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

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

	id      int64
	user    string
	isbn    string
	authors []string
	title   string
	notes   []string
}

type userctx struct {
	state string
	nameFilled bool
	notesFilled bool
	copiesFilled bool

	name string
	notes []string
	copies []int64
}

func (c copyctx) String() string {
	s :=  fmt.Sprintf("id: %v\nuser: %s\nisbn: %s\nauthors: %v\ntitle: %s\nnotes:\n",
		c.id, c.user, c.isbn, c.authors, c.title)
	for _, n := range c.notes {
		s += "\t" + n + "\n"
	}
	return s
}

func (u userctx) String() string {
	s :=  fmt.Sprintf("name: %s\nnotes:\n", u.name)
	for _, n := range u.notes {
		s += "\t" + n + "\n"
	}
	s += fmt.Sprintf("copies: %v\n", u.copies)
	return s
}

func handleCopy(atom sexps.Sexp, parent sexps.Sexp, data interface{}) {
	c := data.(*copyctx)

	switch c.state {
	case "":
		if c.idFilled && c.userFilled && c.isbnFilled && c.authorsFilled && c.titleFilled && c.notesFilled {
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
		if u.nameFilled && u.notesFilled && u.copiesFilled {
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

	case "copies":
		scopies := sexps.List(parent)
		for _, s := range scopies {
			i, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				u.state = "err"
			}

			u.copies = append(u.copies, i)
		}
		u.copiesFilled = true
		u.state = ""

	case "end":
		return

	case "err":
	default:
		fmt.Println("bad state: " + u.state)
	}
}

func main() {
	fnames := []string{"copies", "users"}

	for _, fname := range fnames {
		input, err := os.Open(fname)
		if err != nil {
			fmt.Println(err)
			return
		}
	
		// slurp whole file
		buf, err := ioutil.ReadAll(input)
		if err != nil {
			fmt.Println(err)
			return
		}
		input.Close()
	
		tail := string(buf)
		for len(tail) > 1 { // there's a lonely newline never parsed
			var sexp sexps.Sexp
			var err error
			sexp, tail, err = sexps.Parse(tail)
			if err != nil {
				fmt.Println(err)
				return
			}
	
			fmt.Println(len(tail))

			switch fname {
			case "copies":
				c := copyctx{}
				sexps.Apply(sexp, handleCopy, &c)
				fmt.Println(c)
			case "users":
				u := userctx{}
				sexps.Apply(sexp, handleUser, &u)
				fmt.Println(u)
			}
		}
	}
}
