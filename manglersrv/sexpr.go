package main

import (
	"unicode"
	"unicode/utf8"
)

// atom:   op
// list:   (op arg0 arg1 arg2)
// nested: (op (op2 foo bar) (op3 quux))
type sexpr struct {
	operator string
	args     []sexpr
}

func Parse(s string) *sexpr {
	return nil
}

// token types: "(", ")", atom, quoted atom (string)
func tok(s string) (tok string, tail string) {
	start := -1   // of current run
	str := false

	for i, r := range s {
		switch {
		case r == '(':
			return "(", s[utf8.RuneLen(r):]

		case r == ')':
			return ")", s[utf8.RuneLen(r):]

		case r == '"':
			if !str && start < 0 {
				str = true
				start = i+1
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
