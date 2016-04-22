package main

import (
	"errors"
	"fmt"
)

// Selections, elements and commands working on selections

// selFn describes a selector command (one of .0BCU) which selects a subset of sel.
// The subset is the union of elements that match one of the args.
// XXX Define matching in detail (see offline chart).
type selFn func(sel []elem, args []string) ([]elem, error)

// Interface elem is implemented by books, copies and users and
// contains methods applicable to all of them.
type elem interface {
	fmt.Stringer      // returns the id (copies), ISBN (books) or name (users)
	Print() string    // cmd p (all info)
	Note(note string) // cmd n
	Delete()          // cmd d
}

var seltab = map[rune]selFn{
	'.': func(sel []elem, args []string) ([]elem, error) {
		return sel, nil
	},
	'0': func(sel []elem, args []string) ([]elem, error) {
		return nil, nil
	},
	'B': func(sel []elem, args []string) ([]elem, error) {
		var rsel []elem // returned selection

		for _, s := range args {
			if isISBN13(s) {
				for _, b := range books {
					if s == string(b.isbn) {
						rsel = append(rsel, &b)
					}
				}
			} else {
				return rsel, errors.New("unimplemented selection argument type")

			}

		}

		return rsel, nil
	},
}
