package main

import (
	"fmt"
	"log"
	"os"

	"github.com/okitec/libmangler/elem"
)

// store: save all data in three files, one for users, one for books, one for copies
// in the same s-expr format as in the protocol. The files are truncated at the beginning.
func store(booksFname, copiesFname, usersFname string) {
	var dot []elem.Elem // dummy

	books, err := os.Create(booksFname)
	if err != nil {
		log.Fatalln(err)
	}
	defer books.Close()

	copies, err := os.Create(copiesFname)
	if err != nil {
		log.Fatalln(err)
	}
	defer copies.Close()

	users, err := os.Create(usersFname)
	if err != nil {
		log.Fatalln(err)
	}
	defer users.Close()

	fmt.Fprint(books, interpret("Bp", &dot))
	fmt.Fprint(copies, interpret("Cp", &dot))
	fmt.Fprint(users, interpret("Up", &dot))
}
