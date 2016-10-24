package elems

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

type Copy struct {
	ID    int64
	User  *User
	Book  *Book
	Notes []string
}

// The map Copies holds pointers to all copies indexed by id.
var Copies map[int64]*Copy

func (c *Copy) String() string {
	return fmt.Sprint(c.ID)
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
	if c.User == nil {
		usernil = true
		c.User = &User{"", nil, nil}
	}

	s := fmt.Sprintf(fmtstr, c.ID, c.User, c.Book, strings.Join(c.Book.Authors, `" "`),
		c.Book.Title, strings.Join(c.Notes, "\"\n\t\t\""))

	if usernil {
		c.User = nil
	}

	return s
}

func (c *Copy) Note(note string) {
	c.Notes = append(c.Notes, fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), note))
}

func (c *Copy) Delete() {
	delete(Copies, c.ID)

resized0:
	if c.User != nil {
		for i := range c.User.Copies {
			if c.User.Copies[i] == c {
				// See https://github.com/golang/go/wiki/SliceTricks
				n := len(c.User.Copies)
				c.User.Copies[i] = c.User.Copies[n-1]
				c.User.Copies[n-1] = nil
				c.User.Copies = c.User.Copies[:n-1]
				// reset range loop because slice is shorter
				goto resized0
			}
		}
	}

resized1:
	for i := range c.Book.Copies {
		if c.Book.Copies[i] == c {
			n := len(c.Book.Copies)
			c.Book.Copies[i] = c.Book.Copies[n-1]
			c.Book.Copies[n-1] = nil
			c.Book.Copies = c.Book.Copies[:n-1]
			goto resized1
		}
	}
}

// The method Lend lends a Copy to a User. An error is returned if the Copy is already lent or
// u is nil.
func (c *Copy) Lend(u *User) error {
	if c.User != nil {
		return fmt.Errorf("can't lend %v: already lent", c)
	}

	if u == nil {
		return fmt.Errorf("can't lend %v: no user specified", c)
	}

	c.User = u
	u.Copies = append(u.Copies, c)
	c.Note(fmt.Sprintf("lent to %s", u))
	return nil
}

// The method Return returns a Copy that was lent to an user.
func (c *Copy) Return() {
	u := c.User
	if c.User == nil {
		return
	}

resized:
	for i := range c.User.Copies {
		if c.User.Copies[i] == c {
			n := len(c.User.Copies)
			c.User.Copies[i] = c.User.Copies[n-1]
			c.User.Copies[n-1] = nil
			c.User.Copies = c.User.Copies[:n-1]
			goto resized
		}
	}

	c.User = nil
	c.Note(fmt.Sprintf("returned from %s", u))
}

func NewCopy(id int64, b *Book) (*Copy, error) {
	c := Copy{id, nil, b, nil}

	if b == nil {
		return nil, fmt.Errorf("NewCopy: book doesn't exist")
	}

	if Copies[id] != nil {
		return Copies[id], fmt.Errorf("NewCopy: copy %v already exists", id)
	}

	b.Copies = append(b.Copies, &c)
	Copies[c.ID] = &c
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
