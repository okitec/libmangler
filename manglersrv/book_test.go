package main

import "testing"

var isISBN13Tests = []struct {
	s    string
	pass bool
}{
	{"9780201079814", true},
	{"978-0-201-07981-4", true},

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
