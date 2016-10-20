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

	// To fix issue #9: the string for no user should be "", not the "<nil>"
	// generated when *printf encounters a nil object. So make a dummy.
	var usernil bool
	if c.user == nil {
		usernil = true
		c.user = &User{"", nil, nil}
	}

	s := fmt.Sprintf(fmtstr, c.id, c.user, c.book, strings.Join(c.book.authors, `" "`),
		c.book.title, strings.Join(c.notes, "\"\n\t\t\""))

	if usernil {
		c.user = nil
	}

	return s
}

func (c *Copy) Note(note string) {
	c.notes = append(c.notes, fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), note))
}

func (c *Copy) Delete() {
	delete(copies, c.id)

resized0:
	for i := range c.user.copies {
		if c.user.copies[i] == c {
			// See https://github.com/golang/go/wiki/SliceTricks
			n := len(c.user.copies)
			c.user.copies[i] = c.user.copies[n-1]
			c.user.copies[n-1] = nil
			c.user.copies = c.user.copies[:n-1]
			// reset range loop because slice is shorter
			goto resized0
		}
	}

resized1:
	for i := range c.book.copies {
		if c.book.copies[i] == c {
			n := len(c.user.copies)
			c.user.copies[i] = c.user.copies[n-1]
			c.user.copies[n-1] = nil
			c.user.copies = c.user.copies[:n-1]
			goto resized1
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

resized:
	for i := range c.user.copies {
		if c.user.copies[i] == c {
			n := len(c.user.copies)
			c.user.copies[i] = c.user.copies[n-1]
			c.user.copies[n-1] = nil
			c.user.copies = c.user.copies[:n-1]
			goto resized
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
