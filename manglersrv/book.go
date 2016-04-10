package main

import "fmt"

type ISBN string

// struct book implements elem. XXX
type book struct {
	isbn  ISBN
	title string
	notes []string
	// copies []bcopy // XXX
}

// books contains all book structs.
var books map[ISBN]book

func (b book) String() string {
	return string(b.isbn)
}

func (b book) Print() string {
	return fmt.Sprintln(b.isbn, b.title, b.notes) // XXX implement format as in spec
}

func (b book) Note(note string) {
	b.notes = append(b.notes, note)
}

func (b book) Delete() {
	// XXX delete from disk
	// XXX should this return an error?
	delete(books, b.isbn)
}
