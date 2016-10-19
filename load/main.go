package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/okitec/libmangler/sexps"
)

type context struct {
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

func (ctx context) String() string {
	s :=  fmt.Sprintf("id: %v\nuser: %s\nisbn: %s\nauthors: %v\ntitle: %s\nnotes:\n",
		ctx.id, ctx.user, ctx.isbn, ctx.authors, ctx.title)
	for _, n := range ctx.notes {
		s += "\t" + n + "\n"
	}
	return s
}

func preorder(sexp sexps.Sexp, parent sexps.Sexp, fn func(sexps.Sexp, sexps.Sexp, interface{}), data interface{}) {
	if sexps.IsAtom(sexp) {
		fn(sexp, parent, data)
	}

	if sexp.Car() != nil {
		preorder(sexp.Car(), sexp, fn, data)
	}

	if sexp.Cdr() != nil {
		preorder(sexp.Cdr(), sexp, fn, data)
	}
}

func list(sexp sexps.Sexp) (ls []string) {
	for {
		ls = append(ls, sexp.Car().String())
		sexp = sexp.Cdr()
		if sexp == nil {
			return ls
		}
	}
}

func handle(atom sexps.Sexp, parent sexps.Sexp, data interface{}) {
	ctx := data.(*context)

	switch ctx.state {
	case "":
		if ctx.idFilled && ctx.userFilled && ctx.isbnFilled && ctx.authorsFilled && ctx.titleFilled && ctx.notesFilled {
			ctx.state = "end"
			break
		}

		switch atom.String() {
		case "copy":
			ctx.state = "copy"
		case "user":
			ctx.state = "user"
		case "book":
			ctx.state = "book"
		case "notes":
			ctx.state = "notes"
		default:
			ctx.state = "err"
		}

	case "copy":
		ctx.id, _ = strconv.ParseInt(atom.String(), 10, 64)
		ctx.idFilled = true
		ctx.state = ""

	case "user":
		ctx.user = atom.String()
		ctx.userFilled = true
		ctx.state = ""

	case "book":
		ctx.isbn = atom.String()
		ctx.isbnFilled = true
		ctx.state = "book2"

	case "book2":
		if ctx.authorsFilled && ctx.titleFilled {
			ctx.state = ""
			break
		}

		switch atom.String() {
		case "authors":
			ctx.state = "authors"
		case "title":
			ctx.state = "title"
		default:
			// do nothing; the authors match here
		}

	case "authors":
		ctx.authors = list(parent)
		ctx.authorsFilled = true
		if !ctx.titleFilled {
			ctx.state = "book2"
		} else {
			ctx.state = ""
		}
	case "title":
		ctx.title = atom.String()
		ctx.titleFilled = true
		if !ctx.authorsFilled {
			ctx.state = "book2"
		} else {
			ctx.state = ""
		}

	case "notes":
		ctx.notes = list(parent)
		ctx.notesFilled = true
		ctx.state = ""

	case "end":
		return

	case "err":
	default:
		fmt.Println("bad state: " + ctx.state)
	}
}

func main() {
	var input io.ReadCloser

	fname := "copies"
	if len(os.Args) >= 2 {
		fname = os.Args[1]
	}

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

	ctx := context{}
	sexp := sexps.Parse(string(buf))
	preorder(sexp, nil, handle, &ctx)
	fmt.Println(ctx)

	/*	book := user.Cdr()
		fmt.Println("isbn: " + book.Car().Cdr().Car().String())
			authors := book.Car().Cdr().Cdr()
			fmt.Printf("authors: %q\n", list(authors.Car().Cdr()))
			title := book.Car().Cdr().Cdr().Cdr()
			fmt.Println("title: " + title.Car().Cdr().Car().String())
		notes := book.Cdr()
		fmt.Printf("notes: %q\n", list(notes.Car().Cdr()))
	*/
}
