package main

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

type ISBN string

// Struct Book stores information about books. The physical manifestations of the
// book are called Copies. Books are identified by their unique ISBN.
type Book struct {
	isbn   ISBN
	title  string
	notes  []string
	copies []*Copy
}

// books contains all book structs.
var books map[ISBN]*Book

func (b *Book) String() string {
	return string(b.isbn)
}

// The Print method prints information about the book, including a list of copies,
// in beautifully formatted and indented s-exps.
func (b *Book) Print() string {
	const fmtstr = `(book %q
	(author %s)
	(title %s)
	(notes
		"%s"
	)
	(copies %v)
)`

	return fmt.Sprintf(fmtstr, b.isbn, "WIP", b.title, strings.Join(b.notes, "\"\n\t\t\""), b.copies) // XXX b.copies is printed as [...]
}

// Note saves a note after prepending a ISO 8601 == RFC 3339 date.
func (b *Book) Note(note string) {
	b.notes = append(b.notes, fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), note))
}

func (b *Book) Delete() {
	// XXX delete from disk
	// XXX should this return an error?
	delete(books, b.isbn)
}

// NewBook adds a Book to the system.
// XXX worldcat automated metadata fetching would be nice
func NewBook(isbn, title string) (*Book, error) {
	// XXX check whether Book already exists
	if !isISBN13(isbn) {
		return nil, fmt.Errorf("NewBook: %q is not a ISBN-13")
	}

	b := Book{ISBN(isbn), title, nil, nil}
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
	for i, r := range s {
		// s has more runes than 13 digits + 4 dashes or more digits than expected.
		if i > 17 || ndigits > 13 {
			return false
		}

		if r == '-' {
			continue
		}

		if unicode.IsDigit(r) {
			isbn[ndigits] = int(r) - '0'
			ndigits++
		}
	}

	// Checksum calculation. Note that the odd digits have even indexes in the array.

	sum := 0
	for i := 0; i < 12; i += 2 {
		sum += isbn[i] // odd numbers
	}

	for i := 1; i < 12; i += 2 {
		sum += 3 * isbn[i] // even numbers
	}

	chksum := 10 - sum%10

	if chksum != isbn[12] {
		return false
	}

	return true
}
