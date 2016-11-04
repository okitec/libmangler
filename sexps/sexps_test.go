package sexps

import "testing"

func TestTok(t *testing.T) {
	var tests = []struct {
		s    string
		tok  string // wanted token
		tail string // wanted tail
	}{
		{``, ``, ``},   /* empty */
		{`"`, ``, ``},  /* dangling quote */
		{` `, ``, ` `}, /* lone whitespace */

		{`(`, `(`, ``},
		{`)`, `)`, ``},
		{`derp`, `derp`, ``},         /* atom/number */
		{`""`, ``, ``},               /* empty string */
		{`"foo bar"`, `foo bar`, ``}, /* quoted string with whitespace */

		{`ab"c d"e`, `ab"c`, ` d"e`},  /* embedded quote is auto-escaped */
		{`derp)`, `derp`, `)`},
	}

	for _, tt := range tests {
		tok, tail := tok(tt.s)
		if tok != tt.tok || tail != tt.tail {
			t.Errorf("tok(%q): got (%q, %q), want (%q, %q)\n", tt.s, tok, tail, tt.tok, tt.tail)
		}
	}
}

func TestParse(t *testing.T) {
	var tests = []struct {
		s   string
		want string // expected output in pairwise (x . y) s-exp notation
	}{
		{"derp", "derp"},
		{"(derp)", "(derp . ())"},
		{"(foo bar)", "(foo . (bar . ()))"},
		{"(foo (bar))", "(foo . ((bar . ()) . ()))"},
		{"(foo bar quux derp)", "(foo . (bar . (quux . (derp . ()))))"},
		{"(foo (bar\n(quux derp))))", "(foo . ((bar . ((quux . (derp . ())) . ())) . ()))"},
		{"(foo bar) (quux)", "(foo . (bar . ()))"},  // two sexp; parse only one at a time
		{`"herp derp"`, `"herp derp"`},
		{`("Ἡρόδοτος Ἁλικαρνησσέος")`, `("Ἡρόδοτος Ἁλικαρνησσέος" . ())`},
		{"(* 2 (+ 3 4))", "(* . (2 . ((+ . (3 . (4 . ()))) . ())))"}, // cf. Wikipedia
		{"(A (B C) (D E))", "(A . ((B . (C . ())) . ((D . (E . ())) . ())))"},
		{`(A "" B)`, `(A . ("" . (B . ())))`},
		{`(copy 594
			(user "Dominik Okwieka")
			(book "978..."
				(authors "herp")
				(title "derp")
			)
			(notes "foo")
		)`, `(copy . (594 . ((user . ("Dominik Okwieka" . ())) . ((book . (978... . ((authors . (herp . ())) . ((title . (derp . ())) . ())))) . ((notes . (foo . ())) . ())))))`},

		// XXX add failure cases (dangling braces, ...)
	}

	for _, tt := range tests {
		sxp, tail, err := Parse(tt.s)
		if err != nil {
			t.Errorf("Parse(%q): %v (tail = %s)", tt.s, err, tail)
		}

		got := sxp.Print()
		if got != tt.want {
			t.Errorf("Parse(%q): got %q, want %q\n", tt.s, got, tt.want)
		}
	}
}

func TestList(t *testing.T) {
	var tests = []struct {
		s    string
		want []string
	}{
		{"derp", []string{"derp"}},
		{"(derp)", []string{"derp"}},
		{"(derp herp)", []string{"derp", "herp"}},
		{`(derp "")`, []string{"derp", ""}},
		{`("")`, []string{""}},
	}

	// Initial value in diff for encountered strings. Must be != 0.
	const foo = 100

	for _, tt := range tests {
		sxp, tail, err := Parse(tt.s)
		if err != nil {
			t.Errorf("Parse(%q): %v (tail = %s)", tt.s, err, tail)
		}

		got := List(sxp)
		diff := make(map[string]int)
		for _, s := range got {
			// On first encounter, set start value distinct from 0.
			if _, ok := diff[s]; !ok {
				diff[s] = foo
			}
			diff[s]++
		}

		for _, s := range tt.want {
			// On first encounter, set start value distinct from 0.
			if _, ok := diff[s]; !ok {
				diff[s] = foo
			}
			diff[s]--
		}

		for _, i := range diff {
			if i != foo {
				t.Errorf("List(%q): got %v, want %v; diff is %v\n", tt.s, got, tt.want, i)
			}
		}
	}
}
