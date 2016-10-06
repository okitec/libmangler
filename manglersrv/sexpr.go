package main

// XXX put into separate package

import (
	"fmt"
	"unicode"
)

// s-exp := atom
// s-exp := (x . y) where x and y are s-exps
// To access information, walk the tree and interpret the
// leaves' string values.
type sxp interface {
	fmt.Stringer
	Car() sxp
	Cdr() sxp
}

// So-called cons cell, represented as (car . cdr) where car and cdr
// are also cons cells. car and cdr are the common and archaic names
// for these. Don't question them; they will become natural.
type cell struct {
	car sxp
	cdr sxp
}

// Atoms are non-nested symbols, strings, numbers, etc.
// They are saved as strings, so atoms must be converted
// by the user.
type atom string

func (cp *cell) String() string {
	var scar, scdr string

	if cp.car == nil {
		scar = "()"
	} else {
		scar = cp.car.String()
	}

	if cp.cdr == nil {
		scdr = "()"
	} else {
		scdr = cp.cdr.String()
	}

	return fmt.Sprintf("(%s . %s)", scar, scdr)
}

func (cp *cell) Car() sxp {
	return cp.car
}

func (cp *cell) Cdr() sxp {
	return cp.cdr
}

func (ap *atom) String() string {
	return string(*ap)
}

func (ap *atom) Car() sxp {
	return ap
}

func (ap *atom) Cdr() sxp {
	return nil
}

func cons(car, cdr sxp) *cell {
	return &cell{car, cdr}
}

func mkatom(s string) *atom {
	a := atom(s)
	return &a
}

func Parse(s string) sxp {
	sxp, _ := sexpr(s)
	return sxp
}

// sexpr → atom | ( sexprlist )
func sexpr(s string) (sxp sxp, tail string) {
	t, tail := tok(s)
	switch t {
	case "(":
		sxp, tail = sexprlist(tail)
		t, tail := tok(tail)
		if t != ")" {
			fmt.Println("missing ')'") // XXX return an error
		}

		return sxp, tail

	case ")":
		fmt.Println("unexpected ')'") // XXX return an error
		return nil, tail

	default:
		return mkatom(t), tail
	}
}

// sexprlist → | sexpr sexprlist
func sexprlist(s string) (sxp sxp, tail string) {
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
			} else {
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
	return s + tail
}
