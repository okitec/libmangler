package main

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"
	"unicode"
)

type ISBN string

// Struct Book stores information about books. The physical manifestations of the
// book are called Copies. Books are identified by their unique ISBN.
type Book struct {
	isbn    ISBN
	title   string
	authors []string
	notes   []string
	tags    []string
	copies  []*Copy
}

// books contains all book structs.
var books map[ISBN]*Book

func (b *Book) String() string {
	return string(b.isbn)
}

// The Print method prints information about the book, including a list of copies,
// in beautifully formatted and indented s-exps.
func (b *Book) Print() string {
	const fmtstr = `(book %s
	(authors "%s")
	(title %q)
	(notes
		"%s"
	)
	(tags %s)
	(copies %v)
)`

	return fmt.Sprintf(fmtstr, b.isbn, strings.Join(b.authors, `" "`), b.title,
		strings.Join(b.notes, "\"\n\t\t\""), sTags(b.tags), sCopies(b.copies))
}

// Note saves a note after prepending a ISO 8601 == RFC 3339 date.
func (b *Book) Note(note string) {
	b.notes = append(b.notes, fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), note))
}

func (b *Book) Delete() {
	// XXX should this return an error?

	if len(b.copies) > 0 {
		return
	}

	debug.PrintStack()
	fmt.Printf("delete(%v, %v)\n", books, b.isbn)
	delete(books, b.isbn)
	fmt.Printf("delete(%v, %v)\n", books, b.isbn)
}

func (b *Book) Tag(add bool, tag string) {
	b.tags = addToTags(b.tags, add, tag)
}

// NewBook adds a Book to the system.
func NewBook(isbn, title string, authors []string) (*Book, error) {
	// XXX check whether Book already exists
	if !isISBN13(isbn) {
		return nil, fmt.Errorf("NewBook: %q is not a ISBN-13", isbn)
	}

	b := Book{ISBN(isbn), title, authors, nil, nil, nil}
	books[ISBN(isbn)] = &b
	b.Note("added to the system")
	return &b, nil
}

// The isISBN13 function checks whether its argument is a valid ISBN-13.
// The input string may have dashes between the components. This is not required,
// the check bases solely on length and checksum.
// Examples: "978-0-201-07981-4", "9780201079814"
func isISBN13(s string) bool {
	if len(s) < 13 {
		return false
	}

	var isbn [13]int
	ndigits := 0
	for _, r := range s {
		// s has more digits than expected
		if ndigits >= 13 {
			return false
		}

		if r == '-' {
			continue
		} else if unicode.IsDigit(r) {
			isbn[ndigits] = int(r) - '0'
			ndigits++
		} else {
			return false // bad characters
		}
	}

	// Checksum validation. Note that the odd digits have even indexes in the array.
	// The ISBN is valid if adding the check digit to the sum makes it a multiple of ten.

	sum := 0
	for i := 0; i < 13; i += 2 {
		sum += isbn[i] // odd numbers (includes check digit)
	}

	for i := 1; i < 12; i += 2 {
		sum += 3 * isbn[i] // even numbers
	}

	if (sum % 10) == 0 {
		return true
	}

	return false
}
