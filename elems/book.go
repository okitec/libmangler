package elems

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
	ISBN    ISBN
	Title   string
	Authors []string
	Notes   []string
	Tags    []string
	Copies  []*Copy
}

// Books contains all book structs.
var Books map[ISBN]*Book

func (b *Book) String() string {
	return string(b.ISBN)
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

	return fmt.Sprintf(fmtstr, b.ISBN, strings.Join(b.Authors, `" "`), b.Title,
		strings.Join(b.Notes, "\"\n\t\t\""), sTags(b.Tags), sCopies(b.Copies))
}

// Note saves a note after prepending a ISO 8601 == RFC 3339 date.
func (b *Book) Note(note string) {
	b.Notes = append(b.Notes, fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), note))
}

func (b *Book) Delete() {
	// XXX should this return an error?

	if len(b.Copies) > 0 {
		return
	}

	delete(Books, b.ISBN)
}

func (b *Book) Tag(add bool, tag string) {
	b.tags = addToTags(b.tags, add, tag)
}

// NewBook adds a Book to the system.
func NewBook(isbn, title string, authors []string) (*Book, error) {
	// XXX check whether Book already exists
	if !isISBN13(isbn) {
		return nil, fmt.Errorf("NewBook: %s is not a ISBN-13", isbn)
	}

	b := Book{ISBN(isbn), title, authors, nil, nil}
	Books[ISBN(isbn)] = &b
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
