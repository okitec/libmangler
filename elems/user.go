package elems

import (
	"fmt"
	"strings"
	"time"
)

type User struct {
	Name   string
	Notes  []string
	Tags   []string
	Copies []*Copy
}

var Users map[string]*User

func (u *User) String() string {
	return u.Name
}

func (u *User) Print() string {
	const fmtstr = `(user %q
	(notes
		"%s"
	)
	(tags %s)
	(copies %v)
)`

	return fmt.Sprintf(fmtstr, u.Name, strings.Join(u.Notes, "\"\n\t\t\""), sTags(u.Tags), sCopies(u.Copies))
}

func (u *User) Note(note string) {
	u.Notes = append(u.Notes, fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), note))
}

func (u *User) Delete() {
	delete(Users, u.Name)
}

func (u *User) Tag(add bool, tag string) {
	u.Tags = addToTags(u.Tags, add, tag)
}

// The function NewUser adds a User. If a User of that name already exists,
// it will be returned with an non-nil error.
func NewUser(name string) (*User, error) {
	if Users[name] != nil {
		return Users[name], fmt.Errorf("NewUser: user %q already exists", name)
	}

	u := User{name, nil, nil, nil}
	Users[name] = &u
	return &u, nil
}
