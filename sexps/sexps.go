package sexps

// XXX move into separate Git repo (github.com/okitec/sexps)

import (
	"fmt"
	"strings"
	"unicode"
)

// s-exp := atom
// s-exp := (x . y) where x and y are s-exps
// To access information, walk the tree and interpret the
// leaves' string values.
type Sexp interface {
	fmt.Stringer
	Print() string  // for pretty-printing; like String(), but puts quotes around atoms with spaces
	Car() Sexp
	Cdr() Sexp
}

// So-called cons cell, represented as (car . cdr) where car and cdr
// are also cons cells. car and cdr are the common and archaic names
// for these. Don't question them; they will become natural.
type cell struct {
	car Sexp
	cdr Sexp
}

// Atoms are non-nested symbols, strings, numbers, etc.
// They are saved as strings, so atoms must be converted
// by the user.
type atom string

// Return pretty-printed s-expression; defined to be equal to Print.
func (cp *cell) String() string {
	return cp.Print()
}

// Return pretty-printed s-expression.
func (cp *cell) Print() string {
	var scar, scdr string

	if cp.car == nil {
		scar = "()"
	} else {
		scar = cp.car.Print()
	}

	if cp.cdr == nil {
		scdr = "()"
	} else {
		scdr = cp.cdr.Print()
	}

	return fmt.Sprintf("(%s . %s)", scar, scdr)
}

func (cp *cell) Car() Sexp {
	return cp.car
}

func (cp *cell) Cdr() Sexp {
	return cp.cdr
}

// Return raw atom; even if there are spaces, nothing is quoted.
// The lib user accesses data solely via this function; having
// quotes there would be a hassle.
func (ap *atom) String() string {
	return string(*ap)
}

// Return possibly quoted atom.
func (ap *atom) Print() string {
	s := string(*ap)
	if strings.ContainsAny(s, " \t") {
		return fmt.Sprintf("%q", s)
	}
	return s
}

func (ap *atom) Car() Sexp {
	return nil
}

func (ap *atom) Cdr() Sexp {
	return nil
}

func cons(car, cdr Sexp) *cell {
	return &cell{car, cdr}
}

func mkatom(s string) *atom {
	a := atom(s)
	return &a
}

// IsAtom returns true when sexp is an atom, i.e. is a leaf node.
func IsAtom(sexp Sexp) bool {
	return sexp.Car() == nil && sexp.Cdr() == nil
}

// Parse parses the first s-expression in the string.
func Parse(s string) Sexp {
	sexp, _ := sexpr(s)
	return sexp
}

// sexpr → atom | ( sexprlist )
func sexpr(s string) (sexp Sexp, tail string) {
	t, tail := tok(s)
	switch t {
	case "(":
		sexp, tail = sexprlist(tail)
		t, tail := tok(tail)

		if t != ")" && t != "" {
			fmt.Printf("missing ')'") // XXX return an error
		}

		return sexp, tail

	case ")":
		fmt.Println("unexpected ')'") // XXX return an error
		return nil, tail

	default:
		return mkatom(t), tail
	}
}

// sexprlist → | sexpr sexprlist
func sexprlist(s string) (sexp Sexp, tail string) {
	t, tail := tok(s)
	if t == "" {
		return mkatom(t), tail
	}

	if t == ")" {
		return nil, untok(s, tail)
	}

	car, tail := sexpr(untok(t, tail))
	cdr, tail := sexprlist(tail)
	return cons(car, cdr), tail
}

// token types: "(", ")", atom, quoted atom (string)
func tok(s string) (tok string, tail string) {
	start := -1 // of current run
	str := false

	for i, r := range s {
		switch {
		case r == '(':
			return "(", s[i+1:]

		case r == ')':
			if start >= 0 {
				return s[start:i], s[i:] // s[i:], not i+1, because we need the ')'
			}

			return ")", s[i+1:]

		case r == '"':
			if !str && start < 0 {
				str = true
				start = i + 1
			} else if str {
				return s[start:i], s[i+1:]
			}

		case unicode.IsSpace(r):
			if !str && start >= 0 {
				return s[start:i], s[i:]
			}

		default:
			if start < 0 {
				start = i
			}
		}
	}

	if start < 0 {
		return "", s
	}

	return s[start:], ""
}

func untok(s string, tail string) string {
	// Need to re-add quotes or else we have nuclear fission, which is wrong.
	// (Issue #26) However, don't do that if there are braces in there; they
	// are part of the syntax structure.
	// BAD:  "foo bar") -> foo bar)
	if strings.ContainsAny(s, " \t") && !strings.ContainsAny(s, "()") {
		return fmt.Sprintf("%q", s) + tail
	}
	return s + tail
}
