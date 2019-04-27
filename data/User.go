package data

import (
	"fmt"
	"io/ioutil"
	"path"
)

// User is a collection of data, such as characters, shared storage, or
// otherwise that pertains to a single user.
type User struct {
	Username         string
	Password         string
	Email            string
	loadedCharacters map[string]*Character
}

// loadUser attempts to load a given User from disk and add it to
// the loadedUsers field in Manager.
func (m *Manager) loadUser(user string) (p *User, err error) {
	filepath := path.Join(m.usersPath, user+".user")
	//
	_, readErr := ioutil.ReadFile(filepath)
	if readErr != nil {
		err = &userError{err: readErr.Error()}
		return
	}
	// TODO: Parse it.

	err = &userError{errType: NoSuchUser}
	return
}

// GetUser returns the User tied to a user and pass.
func (m *Manager) GetUser(user string) (u *User, err error) {
	var ok bool
	if u, ok = m.loadedUsers[user]; !ok {
		u, err = m.loadUser(user)
	}
	return
}

// loadUserCharacter attempts to load a given Character from disk and add it
// to the given User's loadedCharacters field.
func (m *Manager) loadUserCharacter(u *User, name string) (c *Character, err error) {
	err = &userError{errType: NoSuchCharacter, err: name}
	return
}

// GetUserCharacter returns the named Character of a given user.
func (m *Manager) GetUserCharacter(u *User, name string) (c *Character, err error) {
	var ok bool
	if c, ok = u.loadedCharacters[name]; !ok {
		c, err = m.loadUserCharacter(u, name)
	}
	return
}

// Errors for User access
const (
	_ = iota
	NoSuchUser
	BadPassword
	NoSuchCharacter
	AccessError
)

type userError struct {
	err     string
	errType int
}

func (e *userError) Error() string {
	switch e.errType {
	case NoSuchUser:
		return fmt.Sprintf("no such user: %s", e.err)
	case BadPassword:
		return fmt.Sprintf("bad password: %s", e.err)
	case NoSuchCharacter:
		return fmt.Sprintf("no such character: %s", e.err)
	case AccessError:
		return fmt.Sprintf("access error: %s", e.err)
	}
	return fmt.Sprintf("undefined error: %s", e.err)
}
