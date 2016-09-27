package main

import (
	"fmt"
	"testing"
)

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
	var tests = []string {
		"derp",
		"(derp)",
		"(foo bar)",
		"(foo (bar))",
		"(foo bar quux derp)",
		"(foo (bar (quux derp))))",
	}

	for _, tt := range tests {
		sexpr := Parse(tt)
		// can't really compare without normalising the output
		fmt.Printf("Parse(%q): got %q\n", tt, sexpr)
	}
}
