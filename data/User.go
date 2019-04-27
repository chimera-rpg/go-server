package data

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// User is a collection of data for a single user, such as characters, shared
// storage, or otherwise
type User struct {
	Username         string
	Password         string
	Email            string
	loadedCharacters map[string]*Character
}

func (m *Manager) checkUser(user string) bool {
	filePath := path.Join(m.usersPath, user+".user")

	if _, err := os.Stat(filePath); err != nil {
		return false
	}
	return true
}

func (m *Manager) writeUser(u *User) (err error) {
	var file *os.File
	filePath := path.Join(m.usersPath, u.Username+".user")

	if file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0600); err != nil {
		err = &userError{err: err.Error()}
		return
	}
	// We really should do some more intelligent serializing, especially for
	// future functionality of shared inventories.
	file.WriteString(fmt.Sprintf("Username %s\nPassword %s\nEmail %s", u.Username, u.Password, u.Email))
	file.Close()
	return
}

// createUser will attempt to create a new user with the given username,
// password, and email.
func (m *Manager) createUser(user string, pass string, email string) (err error) {
	if m.checkUser(user) {
		return &userError{errType: UserExists}
	}
	u := &User{
		Username: user,
		Password: pass,
		Email:    email,
	}
	if err = m.writeUser(u); err != nil {
		err = &userError{err: err.Error()}
	}
	return
}

// loadUser attempts to load a given User from disk and add it to
// the loadedUsers field in Manager.
func (m *Manager) loadUser(user string) (p *User, err error) {
	filepath := path.Join(m.usersPath, user+".user")
	//
	if _, err = ioutil.ReadFile(filepath); err != nil {
		err = &userError{err: err.Error()}
		return
	}
	// TODO: Deserialize it.

	err = &userError{errType: NoSuchUser}
	return
}

// GetUser returns the User tied to a user and pass.
func (m *Manager) GetUser(user string) (u *User, err error) {
	var ok bool
	if u, ok = m.loadedUsers[user]; !ok {
		u, err = m.loadUser(user)
		if err == nil {
			m.loadedUsers[user] = u
		}
	}
	return
}

// loadUserCharacter attempts to load a given Character from disk and add it
// to the given User's loadedCharacters field.
func (m *Manager) loadUserCharacter(u *User, name string) (c *Character, err error) {
	filepath := path.Join(m.usersPath, u.Username+".user", name+".arch")
	//
	if _, err = ioutil.ReadFile(filepath); err != nil {
		err = &userError{err: err.Error()}
		return
	}

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
	UserExists
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
	case UserExists:
		return fmt.Sprintf("user exists: %s", e.err)
	}
	return fmt.Sprintf("undefined error: %s", e.err)
}
