package data

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// User is a collection of data for a single user, such as characters, shared
// storage, or otherwise
type User struct {
	Username         string
	Password         string
	Email            string
	loadedCharacters map[string]*Character
	hasChanges       bool // if there are changes needing to be saved.
}

// CheckUser checks to see if a user file exists.
func (m *Manager) CheckUser(user string) (exists bool, err error) {
	filePath := path.Join(m.usersPath, user+".user")

	if _, serr := os.Stat(filePath); serr != nil {
		if os.IsNotExist(serr) {
			exists = false
			err = &userError{errType: UserNotExists, err: user}
		} else {
			err = &userError{err: err.Error()}
		}
		return
	}
	exists = true
	return
}

// CheckUserPassword returns if the provided plaintext password matches the user's stored password.
func (m *Manager) CheckUserPassword(u *User, password string) (match bool, err error) {
	match, err = comparePasswordAndHash(password, u.Password)
	if err != nil {
		err = &userError{errType: BadPassword, err: err.Error()}
	}
	return
}

func (m *Manager) writeUser(u *User) (err error) {
	if !u.hasChanges {
		return
	}
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

// CreateUser will attempt to create a new user with the given username,
// password, and email.
func (m *Manager) CreateUser(user string, pass string, email string) (err error) {
	if exists, _ := m.CheckUser(user); exists {
		return &userError{errType: UserExists}
	}
	encodedHash, err := encodePassword(pass, &m.cryptParams)
	if err != nil {
		return err
	}
	u := &User{
		Username:   user,
		Password:   encodedHash,
		Email:      email,
		hasChanges: true,
	}
	if err = m.writeUser(u); err != nil {
		err = &userError{err: err.Error()}
	}
	return
}

// loadUser attempts to load a given User from disk and add it to
// the loadedUsers field in Manager.
func (m *Manager) loadUser(user string) (u *User, err error) {
	var exists bool
	var bytes []byte
	filePath := path.Join(m.usersPath, user+".user")

	if exists, err = m.CheckUser(user); !exists || err != nil {
		return
	}
	//
	if bytes, err = ioutil.ReadFile(filePath); err != nil {
		err = &userError{err: err.Error()}
		return
	}
	u = &User{
		loadedCharacters: make(map[string]*Character),
	}
	// NOTE: For now we're not implementing a parser as it'd be too
	// heavy for the functionality we need at the moment.
	lines := strings.Split(string(bytes), "\n")
	for _, line := range lines {
		ws := strings.Index(line, " ")
		if ws == -1 {
			continue
		}
		key := line[:ws]
		value := line[ws+1:]
		switch key {
		case "Username":
			u.Username = value
		case "Password":
			u.Password = value
		case "Email":
			u.Email = value
		}
	}
	if u.Username == "" {
		err = &userError{errType: BadData, err: "Username missing"}
	}
	if u.Password == "" {
		err = &userError{errType: BadData, err: "Password missing"}
	}
	if err != nil {
		u = nil
	}

	return
}

func (m *Manager) unloadUser(user string) (err error) {
	if u, ok := m.loadedUsers[user]; ok {
		err = m.writeUser(u)
		delete(m.loadedUsers, user)
	}
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
	filepath := path.Join(m.usersPath, u.Username, name+".arch")
	fmt.Printf("Loading character \"%s\"", filepath)
	//
	r, err := ioutil.ReadFile(filepath)
	if err != nil {
		err = &userError{err: err.Error()}
		return
	}

	archetypesMap := make(map[string]Archetype)

	if err = yaml.Unmarshal(r, &archetypesMap); err != nil {
		err = &userError{err: err.Error()}
		return
	}

	var targetArch *Archetype
	for k, archetype := range archetypesMap {
		if k == name {
			targetArch = &archetype
		}
	}

	if targetArch == nil {
		err = &userError{errType: NoSuchCharacter, err: name}
	}

	c = &Character{
		Archetype: targetArch,
	}
	fmt.Println("Looks good...")

	return
}

// GetUserCharacter returns the named Character of a given user.
func (m *Manager) GetUserCharacter(u *User, name string) (c *Character, err error) {
	var ok bool
	if c, ok = u.loadedCharacters[name]; !ok {
		c, err = m.loadUserCharacter(u, name)
		if err == nil {
			u.loadedCharacters[name] = c
		}
	}
	return
}

// Errors for User access
const (
	_ = iota
	NoSuchUser
	BadPassword
	NoSuchCharacter
	BadData
	AccessError
	UserExists
	UserNotExists
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
	case UserNotExists:
		return fmt.Sprintf("user does not exist: %s", e.err)
	}
	return fmt.Sprintf("undefined error: %s", e.err)
}
