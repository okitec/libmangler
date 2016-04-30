package main

import "testing"

var isISBN13Tests = []struct {
	s    string
	pass bool
}{
	{"9780201079814", true},
	{"978-0-201-07981-4", true},
	{"978-3-468-11032-0", true},
	{"978-81-203-0596-0", true},

	{"", false},
	{"abcdefghijklmn", false},
	{"978-0-201-07981-9", false}, // bad checksum
	{"3945856", false},
	{"0-201-07981-X", false}, // ISBN-10
	{"00000000000000000000000000", false}, // too long
}

func TestIsISBN13(t *testing.T) {
	for _, tt := range isISBN13Tests {
		p := isISBN13(tt.s)
		if p != tt.pass {
			t.Errorf("isISBN13(%q): got %v, want %v\n", tt.s, p, tt.pass)
		}
	}
}

func TestBook_String(t *testing.T) {
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
	b := eNewBook(t, "978-0-201-07981-4", "The AWK Programming Language", []string{"Alfred V. Aho", "Brian W. Kernighan", "Peter J. Weinberger"})

	c, _ := NewCopy(b)
	b.Delete()
	_, ok := books[ISBN("978-0-201-07981-4")]
	if !ok {
		t.Fatalf("Book.Delete: deleted even though copies exist\n")
	}

	c.Delete()
	b.Delete()
	b2, ok := books[ISBN("978-0-201-07981-4")]
	if ok {
		t.Fatalf("Book.Delete: still in books map after Delete (b2 = %v)\n", b2)
	}
}

func eNewBook(t *testing.T, isbn, title string, authors []string) *Book {
	// Nicely reset the global tables beforehand.
	books = make(map[ISBN]*Book)
	copies = make(map[int64]*Copy)

	b, err := NewBook("978-0-201-07981-4", "The AWK Programming Language", nil)
	if err != nil {
		t.Fatalf("can't create example book: %v", err)
	}
	return b
}
