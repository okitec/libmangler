package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Copy struct {
	id    int
	user  *User
	book  *Book
	notes []string
}

// The slice copies holds pointers to all copied indexed by id.
// Unlike users and books, a slice suffices here.
var copies []*Copy

func (c *Copy) String() string {
	return fmt.Sprint(c.id)
}

func (c *Copy) Print() string {
	const fmtstr = `(copy %v
	(user %q)
	(book %q
		(author %s)
		(title %q)
	)
	(notes
		"%s"
	)
)
`

	return fmt.Sprintf(fmtstr, c.id, c.user, c.book, "WIP", c.book.title, strings.Join(c.notes, "\"\n\t\t\""))
}

func (c *Copy) Note(note string) {
	c.notes = append(c.notes, fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), note))
}

func (c *Copy) Delete() {
	// XXX This doesn't compress the slices.
	copies[c.id] = nil
	c.user.copies[c.id] = nil
	c.book.copies[c.id] = nil
}

func NewCopy(b *Book) (*Copy, error) {
	// XXX collisions may happen; FIX!
	c := Copy{rand.Int(), nil, b, nil}
	b.copies = append(b.copies, &c)
	return &c, nil
}
