package main

import (
	"fmt"
	"log"
	"os"
)

// store: save all data in three files, one for users, one for books, one for copies
// in the same s-expr format as in the protocol. The files are truncated at the beginning.
func store() {
	var dot []elem  // dummy

	users, err := os.Create("users")
	if err != nil {
		log.Panicln("can't create users file: ", err)
	}
	defer users.Close()

	books, err := os.Create("books")
	if err != nil {
		log.Panicln("can't create users file: ", err)
	}
	defer books.Close()

	copies, err := os.Create("copies")
	if err != nil {
		log.Panicln("can't create users file: ", err)
	}
	defer copies.Close()

	fmt.Fprint(users, interpret("Up", &dot))
	fmt.Fprint(books, interpret("Bp", &dot))
	fmt.Fprint(copies, interpret("Cp", &dot))
}
