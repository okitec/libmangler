package main

import "testing"

func TestTok(t *testing.T) {
	var tests = []struct {
		s    string
		tok  string  // wanted token
		tail string  // wanted tail
	}{
		{``, ``, ``},                   /* empty */
		{`"`, ``, ``},                  /* dangling quote */
		{` `, ``, ` `},                 /* lone whitespace */

		{`(`, `(`, ``},
		{`)`, `)`, ``},
		{`derp`, `derp`, ``},           /* atom/number */
		{`""`, ``, ``},                 /* empty string */
		{`"foo bar"`, `foo bar`, ``},   /* quoted string with whitespace */

		{`ab"c d"e`, `ab`, `c d"e`},    /* embedded quote has no effect */
	}

	for _, tt := range tests {
		tok, tail := tok(tt.s)
		if tok != tt.tok || tail != tt.tail {
			t.Errorf("tok(%q): got (%q, %q), want (%q, %q)\n", tt.s, tok, tail, tt.tok, tt.tail)
		}
	}
}
