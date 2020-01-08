package data

import (
	"strconv"
)

// Variable is the interface that represents strings, ints, or otherwise.
type Variable interface {
	isVariable()
	GetString() (string, error)
	GetInt() (int, error)
}

// Int is an integer type.
type Int int

func (i Int) isVariable() {}

// GetString returns the Int as a string value.
func (i Int) GetString() (string, error) {
	return strconv.Itoa(int(i)), nil
}

// GetInt returns the Int's value.
func (i Int) GetInt() (int, error) {
	return int(i), nil
}

// Bool is our bool type.
type Bool bool

func (b Bool) isVariable() {}

// GetInt returns 0 or 1 of a bool.
func (b Bool) GetInt() (int, error) {
	if b == true {
		return 1, nil
	}
	return 0, nil
}

// GetString returns "true" or "false".
func (b Bool) GetString() (string, error) {
	if b == true {
		return "true", nil
	}
	return "false", nil
}

// String is our string type.
type String string

func (s String) isVariable() {}

// GetString returns the string's value.
func (s String) GetString() (string, error) {
	return string(s), nil
}

// GetInt attempts to return the string as an integer value.
func (s String) GetInt() (int, error) {
	return strconv.Atoi(string(s))
}

// Expression is a more complex expression used for calculating a number.
type Expression string

func (e Expression) isVariable() {}

// GetString doesn't do anything.
func (e Expression) GetString() (string, error) {
	return string(e), nil
}

// GetInt doesn't do anything.
func (e Expression) GetInt() (int, error) {
	return 0, nil
}

// Container is a Variable that is a map of strings to Variables.
type Container map[string]Variable

func (c Container) isVariable() {}
