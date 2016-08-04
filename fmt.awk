# fmt.awk: format into Linelen-character lines
# (code copied from Rob Pike's and Brian W. Kernighan's "The Practice of Programming",
# p. 229, to avoid reinventing the wheel.)

BEGIN { Linelen = 80 }

/./  { for(i = 1; i <= NF; i++) addword($i)}  # non-blank line
/^$/ { printline(); print "" }                # blank line
END  { printline() }

function addword(w) {
	if(length(line) + 1 + length(w) > Linelen)
		printline()
	if(length(line) == 0)
		line = w
	else
		line = line " " w
}

function printline() {
	if(length(line) > 0) {
		print line
		line = ""
	}
}
