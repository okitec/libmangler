package elem

import (
	"fmt"
	"strconv"
)

// Selections, elements and commands working on selections

// selFn describes a selector command (one of .0BCU) which selects a subset of sel.
// The subset is the union of elements that match one of the args.
// XXX Define matching in detail (see offline chart).
type selFn func(sel []Elem, args []string) ([]Elem, error)

// Interface elem is implemented by Books, Copies and Users and
// contains methods applicable to all of them.
type Elem interface {
	fmt.Stringer              // returns the id (Copies), ISBN (Books) or name (Users)
	Print() string            // cmd p (all info)
	Note(note string)         // cmd n  // XXX make fmt-like
	Delete()                  // cmd d
	Tag(add bool, tag string) // cmd t
}

var Books map[ISBN]*Book
var Copies map[int64]*Copy
var Users map[string]*User

// Called during package initialisation and thus before any main().
func init() {
	Books = make(map[ISBN]*Book)
	Users = make(map[string]*User)
	Copies = make(map[int64]*Copy)
}

func Select(r rune, sel []Elem, args []string) ([]Elem, error) {
	fn := seltab[r]
	if fn == nil {
		return sel, fmt.Errorf("invalid selector %q", r)
	}

	return fn(sel, args)
}

// XXX Write a script that generates this from a table. This is too mechanical.
var seltab = map[rune]selFn{
	'.': func(sel []Elem, args []string) ([]Elem, error) {
		return sel, nil
	},
	'0': func(sel []Elem, args []string) ([]Elem, error) {
		return nil, nil
	},
	'B': func(sel []Elem, args []string) ([]Elem, error) {
		var rsel []Elem // returned selection

		// Select all if no constraints given.
		if args == nil {
			for _, b := range Books {
				rsel = append(rsel, b)
			}

			return rsel, nil
		}

		for _, s := range args {
			if isISBN13(s) {
				for _, b := range Books {
					if s == string(b.ISBN) {
						rsel = append(rsel, b)
					}
				}
			} else if id, err := strconv.ParseInt(s, 10, 64); err == nil {
				for _, b := range Books {
					for _, c := range b.Copies {
						if c.ID == id {
							rsel = append(rsel, b)
						}
					}
				}
			} else if s[0] == '#' {
				for _, b := range Books {
					for _, t := range b.Tags {
						if t == s {
							rsel = append(rsel, b)
						}
					}
				}
			} else {
				for _, b := range Books {
					for _, c := range b.Copies {
						if c.User.Name == s {
							rsel = append(rsel, b)
						}
					}
				}
			}

		}

		return rsel, nil
	},
	'C': func(sel []Elem, args []string) ([]Elem, error) {
		var rsel []Elem

		// Select all if no constraints given.
		if args == nil {
			for _, c := range Copies {
				rsel = append(rsel, c)
			}

			return rsel, nil
		}

		for _, s := range args {
			if isISBN13(s) {
				for _, b := range Books {
					if s == string(b.ISBN) {
						// Convert from []*Copy to []Elem
						var cs []Elem
						for _, c := range b.Copies {
							cs = append(cs, c)
						}
						rsel = append(rsel, cs...)
					}
				}
			} else if id, err := strconv.ParseInt(s, 10, 64); err == nil {
				for _, c := range Copies {
					if c.ID == id {
						rsel = append(rsel, c)
					}
				}
			} else if s[0] == '#' {
				for _, c := range Copies {
					for _, t := range c.Tags {
						if t == s {
							rsel = append(rsel, c)
						}
					}
				}
			} else {
				for _, u := range Users {
					if u.Name == s {
						// Convert from []*Copy to []Elem
						var cs []Elem
						for _, c := range u.Copies {
							cs = append(cs, c)
						}
						rsel = append(rsel, cs...)
					}
				}
			}
		}

		return rsel, nil
	},
	'U': func(sel []Elem, args []string) ([]Elem, error) {
		var rsel []Elem

		// Select all if no constraints given.
		if args == nil {
			for _, u := range Users {
				rsel = append(rsel, u)
			}

			return rsel, nil
		}

		for _, s := range args {
			if isISBN13(s) {
				// Do it this way around to avoid duplicate Users.
				for _, u := range Users {
					for _, c := range u.Copies {
						if s == string(c.Book.ISBN) {
							rsel = append(rsel, u)
						}
					}
				}
			} else if id, err := strconv.ParseInt(s, 10, 64); err == nil {
				for _, u := range Users {
					for _, c := range u.Copies {
						if c.ID == id {
							rsel = append(rsel, u)
						}
					}
				}
			} else if s[0] == '#' {
				for _, u := range Users {
					for _, t := range u.Tags {
						if t == s {
							rsel = append(rsel, u)
						}
					}
				}
			} else {
				for _, u := range Users {
					if s == u.Name {
						rsel = append(rsel, u)
					}
				}
			}
		}

		return rsel, nil
	},
}
