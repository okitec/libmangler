package main

import (
	"fmt"
	"strings"
	"time"
)

type User struct {
	name   string
	notes  []string
	tags   []string
	copies []*Copy
}

var users map[string]*User

func (u *User) String() string {
	return u.name
}

func (u *User) Print() string {
	const fmtstr = `(user %q
	(notes
		"%s"
	)
	(tags %s)
	(copies %v)
)`

	return fmt.Sprintf(fmtstr, u.name, strings.Join(u.notes, "\"\n\t\t\""), sTags(u.tags), sCopies(u.copies))
}

func (u *User) Note(note string) {
	u.notes = append(u.notes, fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), note))
}

func (u *User) Delete() {
	delete(users, u.name)
}

func (u *User) Tag(add bool, tag string) {
	u.tags = addToTags(u.tags, add, tag)
}

// The function NewUser adds a User. If a User of that name already exists,
// it will be returned with an non-nil error.
func NewUser(name string) (*User, error) {
	if users[name] != nil {
		return users[name], fmt.Errorf("NewUser: user %q already exists", name)
	}

	u := User{name, nil, nil, nil}
	users[name] = &u
	return &u, nil
}
