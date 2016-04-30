package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Copy struct {
	id    int64
	user  *User
	book  *Book
	notes []string
}

// The map copies holds pointers to all copies indexed by id.
var copies map[int64]*Copy

func (c *Copy) String() string {
	return fmt.Sprint(c.id)
}

func (c *Copy) Print() string {
	const fmtstr = `(copy %v
	(user %q)
	(book %q
		(authors "%s")
		(title %q)
	)
	(notes
		"%s"
	)
)`

	return fmt.Sprintf(fmtstr, c.id, c.user, c.book, strings.Join(c.book.authors, `" "`),
		c.book.title, strings.Join(c.notes, "\"\n\t\t\""))
}

func (c *Copy) Note(note string) {
	c.notes = append(c.notes, fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), note))
}

func (c *Copy) Delete() {
	delete(copies, c.id)

	if c.user != nil {
		for i := range c.user.copies {
			if c.user.copies[i] == c {
				c.user.copies[i] = nil // XXX This doesn't compress the slices.
			}
		}
	}

	for i := range c.book.copies {
		if c.book.copies[i] == c {
			c.book.copies[i] = nil // XXX This doesn't compress the slices.
		}
	}
}

// The method Lend lends a Copy to a User. An error is returned if the Copy is already lent or
// u is nil.
func (c *Copy) Lend(u *User) error {
	if c.user != nil {
		return fmt.Errorf("can't lend %v: already lent", c)
	}

	if u == nil {
		return fmt.Errorf("can't lend %v: no user specified", c)
	}

	c.user = u
	u.copies = append(u.copies, c)
	c.Note(fmt.Sprintf("lent to %s", u))
	return nil
}

// The method Return returns a Copy that was lent to an user.
func (c *Copy) Return() {
	u := c.user
	if c.user == nil {
		return
	}

	for i := range c.user.copies {
		if c.user.copies[i] == c {
			c.user.copies[i] = nil // XXX This doesn't compress the slices.
		}
	}

	c.user = nil
	c.Note(fmt.Sprintf("returned from %s", u))
}

func NewCopy(b *Book) (*Copy, error) {
	// XXX collisions may happen; FIX!
	c := Copy{int64(rand.Int()), nil, b, nil}
	b.copies = append(b.copies, &c)
	copies[c.id] = &c
	return &c, nil
}

// Function sCopies takes an array of string and concatenates it to a space-separated list string.
func sCopies(copies []*Copy) string {
	var buf bytes.Buffer

	for _, c := range copies {
		buf.WriteString(c.String())
		buf.WriteRune(' ')
	}

	return buf.String()
}
