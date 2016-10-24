package main

import (
	"fmt"
	"strconv"
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
	Note(note string) // cmd n  // XXX make fmt-like
	Delete()          // cmd d
}

// XXX Write a script that generates this from a table. This is too mechanical.
var seltab = map[rune]selFn{
	'.': func(sel []elem, args []string) ([]elem, error) {
		return sel, nil
	},
	'0': func(sel []elem, args []string) ([]elem, error) {
		return nil, nil
	},
	'B': func(sel []elem, args []string) ([]elem, error) {
		var rsel []elem // returned selection

		// Select all if no constraints given.
		if args == nil {
			for _, b := range books {
				rsel = append(rsel, b)
			}

			return rsel, nil
		}

		for _, s := range args {
			if isISBN13(s) {
				for _, b := range books {
					if s == string(b.isbn) {
						rsel = append(rsel, b)
					}
				}
			} else if id, err := strconv.ParseInt(s, 10, 64); err == nil {
				for _, b := range books {
					for _, c := range b.copies {
						if c.id == id {
							rsel = append(rsel, b)
						}
					}
				}
			} else {
				for _, b := range books {
					for _, c := range b.copies {
						if c.user.name == s {
							rsel = append(rsel, b)
						}
					}
				}
			}

		}

		return rsel, nil
	},
	'C': func(sel []elem, args []string) ([]elem, error) {
		var rsel []elem

		// Select all if no constraints given.
		if args == nil {
			for _, c := range copies {
				rsel = append(rsel, c)
			}

			return rsel, nil
		}

		for _, s := range args {
			if isISBN13(s) {
				for _, b := range books {
					if s == string(b.isbn) {
						// Convert from []*Copy to []elem
						var cs []elem
						for _, c := range b.copies {
							cs = append(cs, c)
						}
						rsel = append(rsel, cs...)
					}
				}
			} else if id, err := strconv.ParseInt(s, 10, 64); err == nil {
				for _, c := range copies {
					if c.id == id {
						rsel = append(rsel, c)
					}
				}
			} else {
				for _, u := range users {
					if u.name == s {
						// Convert from []*Copy to []elem
						var cs []elem
						for _, c := range u.copies {
							cs = append(cs, c)
						}
						rsel = append(rsel, cs...)
					}
				}
			}
		}

		return rsel, nil
	},
	'U': func(sel []elem, args []string) ([]elem, error) {
		var rsel []elem

		// Select all if no constraints given.
		if args == nil {
			for _, u := range users {
				rsel = append(rsel, u)
			}

			return rsel, nil
		}

		for _, s := range args {
			if isISBN13(s) {
				// Do it this way around to avoid duplicate users.
				for _, u := range users {
					for _, c := range u.copies {
						if s == string(c.book.isbn) {
							rsel = append(rsel, u)
						}
					}
				}
			} else if id, err := strconv.ParseInt(s, 10, 64); err == nil {
				for _, u := range users {
					for _, c := range u.copies {
						if c.id == id {
							rsel = append(rsel, u)
						}
					}
				}
			} else {
				for _, u := range users {
					if s == u.name {
						rsel = append(rsel, u)
					}
				}
			}
		}

		return rsel, nil
	},
}
