package data

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"

	"gopkg.in/yaml.v2"
)

// User is a collection of data for a single user, such as characters, shared
// storage, or otherwise
type User struct {
	Username   string                `yaml:"Username"`
	Password   string                `yaml:"Password"`
	Email      string                `yaml:"Email"`
	Characters map[string]*Character `yaml:"Characters"`
	hasChanges bool                  // if there are changes needing to be saved.
	userPath   string                // filepath of the given user file.
	mutex      sync.Mutex
}

// CheckUser checks to see if a user file exists.
func (m *Manager) CheckUser(user string) (exists bool, err error) {
	filePath := path.Join(m.usersPath, user+".user.yaml")

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
	u.mutex.Lock()
	defer u.mutex.Unlock()
	if !u.hasChanges {
		return
	}
	filePath := path.Join(m.usersPath, u.Username+".user.yaml")

	var bytes []byte
	if bytes, err = yaml.Marshal(u); err != nil {
		log.Print(err)
		err = &userError{err: err.Error()}
		return
	}
	if err = ioutil.WriteFile(filePath, bytes, 0644); err != nil {
		log.Print(err)
		err = &userError{err: err.Error()}
		return
	}

	u.hasChanges = false

	return
}

// CreateUser will attempt to create a new user with the given username,
// password, and email.
func (m *Manager) CreateUser(user string, pass string, email string) (err error) {
	m.usersMutex.Lock()
	defer m.usersMutex.Unlock()
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
	log.Printf("Loading user %s\n", user)
	var exists bool
	var bytes []byte
	filePath := path.Join(m.usersPath, user+".user.yaml")

	if exists, err = m.CheckUser(user); !exists || err != nil {
		return
	}

	//
	if bytes, err = ioutil.ReadFile(filePath); err != nil {
		err = &userError{err: err.Error()}
		return
	}

	if err = yaml.Unmarshal(bytes, &u); err != nil {
		log.Print(err)
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

// unloadUser attempts to unload the given user by name.
func (m *Manager) unloadUser(u *User) (err error) {
	err = m.writeUser(u)
	return
}

// GetUser returns the User tied to a user and pass.
func (m *Manager) GetUser(user string) (u *User, err error) {
	m.usersMutex.Lock()
	defer m.usersMutex.Unlock()
	var ok bool
	if u, ok = m.loadedUsers[user]; !ok {
		u, err = m.loadUser(user)
		if err == nil {
			m.loadedUsers[user] = u
		}
	}
	return
}

// CleanupUser unloads the given user by name.
func (m *Manager) CleanupUser(user string) (err error) {
	m.usersMutex.Lock()
	defer m.usersMutex.Unlock()
	log.Printf("Cleaning up user %s\n", user)
	if u, ok := m.loadedUsers[user]; ok {
		err = m.unloadUser(u)
		delete(m.loadedUsers, user)
	}
	return
}

// CreateUserCharacter will attempt to create a new character named by the
// given name.
func (m *Manager) CreateUserCharacter(u *User, name string) (err error) {
	if name == "" {
		return &userError{errType: EmptyCharacterName}
	}
	if exists, _ := m.CheckUserCharacter(u, name); exists {
		return &userError{errType: SuchCharacter, err: name}
	}

	c := &Character{
		Name: name,
		Archetype: Archetype{
			Name: StringExpression{src: name},
		},
		SaveInfo: SaveInfo{
			Map: "Chamber of Origins",
			Y:   0,
			X:   0,
			Z:   0,
		},
		//Race: m.Strings.Lookup(m.pcArchetypes[0].SelfID),
	}

	u.mutex.Lock()
	u.Characters[name] = c
	u.hasChanges = true
	u.mutex.Unlock()

	if err = m.writeUser(u); err != nil {
		err = &userError{err: err.Error()}
	}

	return
}

// CheckUserCharacter checks to see if the given character exists
// for the provided user.
func (m *Manager) CheckUserCharacter(u *User, name string) (exists bool, err error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	if _, ok := u.Characters[name]; ok {
		exists = true
	} else {
		err = &userError{errType: NoSuchCharacter, err: name}
		exists = false
	}

	return
}

// GetUserCharacter returns the given character by name if it eixsts.
func (m *Manager) GetUserCharacter(u *User, name string) (c *Character, err error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	if character, ok := u.Characters[name]; ok {
		c = character
	} else {
		err = &userError{errType: NoSuchCharacter, err: name}
	}

	return
}

// Errors for User access
const (
	_ = iota
	NoSuchUser
	BadPassword
	EmptyCharacterName
	NoSuchCharacter
	SuchCharacter
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
	case EmptyCharacterName:
		return fmt.Sprintf("empty character name")
	case NoSuchCharacter:
		return fmt.Sprintf("no such character: %s", e.err)
	case SuchCharacter:
		return fmt.Sprintf("character exists: %s", e.err)
	case AccessError:
		return fmt.Sprintf("access error: %s", e.err)
	case UserExists:
		return fmt.Sprintf("user exists: %s", e.err)
	case UserNotExists:
		return fmt.Sprintf("user does not exist: %s", e.err)
	}
	return fmt.Sprintf("undefined error: %s", e.err)
}
