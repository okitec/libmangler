package sexp

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

		{`ab"c d"e`, `ab`, `c d"e`},  /* embedded quote has no effect */
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
	}

	for _, tt := range tests {
		sxp := Parse(tt.s)
		got := sxp.String()
		if got != tt.want {
			t.Errorf("Parse(%q): got %q, want %q\n", tt.s, got, tt.want)
		}
	}
}
