package main

import (
	"fmt"
	"testing"
)

func TestIsISBN13(t *testing.T) {
	var tests = []struct {
		s    string
		pass bool
	}{
		{"9780201079814", true},
		{"978-0-201-07981-4", true},
		{"978-3-468-11032-0", true},
		{"978-81-203-0596-0", true},

		{"", false},                           // empty
		{"abcdefghijklmn", false},             // not a number
		{"978-0-201-07981-9", false},          // bad checksum
		{"3945856", false},                    // too short
		{"0-201-07981-X", false},              // ISBN-10
		{"00000000000000000000000000", false}, // too long
	}

	for _, tt := range tests {
		p := isISBN13(tt.s)
		if p != tt.pass {
			t.Errorf("isISBN13(%q): got %v, want %v\n", tt.s, p, tt.pass)
		}
	}
}

func TestBook_String(t *testing.T) {
	books = make(map[ISBN]*Book)

	b := eNewBook(t, "978-0-201-07981-4", "The AWK Programming Language", nil)

	if b.String() != string(b.isbn) {
		t.Errorf("Book.String: got %v, want %v\n", b.String(), string(b.isbn))
	}

	b.Delete()
}

func TestBook_Print(t *testing.T) {
	t.Skip()

	// XXX need a s-exp parser
	//	b := eNewBook(t, "978-0-201-07981-4", "The AWK Programming Language", []string{"Alfred V. Aho", "Brian W. Kernighan", "Peter J. Weinberger"})
	//	s := b.Print()
}

func TestBook_Note(t *testing.T) {
	t.Skip()

	// XXX how to really test?
	//	b := eNewBook(t, "978-0-201-07981-4", "The AWK Programming Language", []string{"Alfred V. Aho", "Brian W. Kernighan", "Peter J. Weinberger"})
	//	b.Note("foobar")
	//	b.Note("quux")
}

func TestBook_Delete(t *testing.T) {
	books = make(map[ISBN]*Book)
	copies = make(map[int64]*Copy)
	b := eNewBook(t, "978-0-201-07981-4", "The AWK Programming Language", []string{"Alfred V. Aho", "Brian W. Kernighan", "Peter J. Weinberger"})

	c, _ := NewCopy(b)
	b.Delete() // XXX not called - what am I missing?
	_, ok := books[ISBN("978-0-201-07981-4")]
	if !ok {
		t.Fatalf("Book.Delete: deleted even though copies exist\n")
	}

	c.Delete()
	b.Delete() // XXX not called - what am I missing?
	fmt.Println("after second b.Delete: books =", books)
	b2, ok := books[ISBN("978-0-201-07981-4")]
	if ok {
		t.Fatalf("Book.Delete: still in books map after Delete\n")
	}
}

func TestNewBook(t *testing.T) {
	books = make(map[ISBN]*Book)
	b := eNewBook(t, "978-0-201-07981-4", "The AWK Programming Language", []string{"Alfred V. Aho", "Brian W. Kernighan", "Peter J. Weinberger"})

	if b == nil {
		t.Fatalf("TestNewBook: b is nil!\n")
	}

	if books[ISBN("978-0-201-07981-4")] != b {
		t.Fatalf("TestNewBook: books[ISBN(\"978-0-201-07981-4\")] is not equal to b\n")
	}

	if b.isbn != ISBN("978-0-201-07981-4") || b.title != "The AWK Programming Language" {
		t.Fatalf("TestNewBook: mangled ISBN or title\n")
	}

	// XXX test authors
}

func eNewBook(t *testing.T, isbn, title string, authors []string) *Book {
	b, err := NewBook("978-0-201-07981-4", "The AWK Programming Language", nil)
	if err != nil {
		t.Fatalf("can't create example book: %v", err)
	}
	return b
}
