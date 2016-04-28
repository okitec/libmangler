package main

import (
	"fmt"
	"strings"
	"time"
)

type User struct {
	name  string
	notes []string
	//copies []Copy
}

var users map[string]*User

func (u *User) String() string {
	return u.name
}

func (u *User) Print() string {
	const fmtstr = `
(user %q
	(notes
		"%s"
	)
	(copies %v)
)`

	return fmt.Sprintf(fmtstr, u.name, strings.Join(u.notes, "\"\n\t\t\""), "WIP")
}

func (u *User) Note(note string) {
	u.notes = append(u.notes, fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), note))
}

func (u *User) Delete() {
	delete(users, u.name)
}
