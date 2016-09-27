package main

import "unicode"

// atom:   op
// list:   (op arg0 arg1 arg2)
// nested: (op (op2 foo bar) (op3 quux))
type Sexpr struct {
	operator string
	args     []Sexpr
}

func Parse(s string) Sexpr {
	sexpr, _ := sexprParse(s)
	return sexpr
}

// Sexpr → ( op Sexpr-list | op 
func sexprParse(s string) (sexpr Sexpr, tail string) {
	t, tail := tok(s)
	if t == "(" {
		t, tail = tok(tail)
		return Sexpr{t, sexprList(tail)}, tail
	}

	return Sexpr{t, nil}, tail
}

// Sexpr-list → ) | Sexpr Sexpr-list )
func sexprList(s string) []Sexpr {
	t, tail := tok(s)
	if t == "" {
		return nil
	}

	sexpr, tail := sexprParse(tail)
	listhead := []Sexpr{sexpr}
	listtail := sexprList(tail)

	if listtail != nil {
		return append(listhead, listtail...)
	}
	return listhead
}

func (sexpr Sexpr) String() string {
	if sexpr.args == nil {
		return sexpr.operator
	}

	s := "("
	s += sexpr.operator
	s += " "

	for _, arg := range sexpr.args {
		s += arg.String() + " "
	}

	s += ")"
	return s
}

// token types: "(", ")", atom, quoted atom (string)
func tok(s string) (tok string, tail string) {
	start := -1   // of current run
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
