package GameData

import (
  "strconv"
)

type VariableType struct {
}

type Variable interface {
  isVariable()
  GetString() (string, error)
  GetInt() (int, error)
}

type Int int
func (i Int) isVariable() {}
func (i Int) GetString() (string, error) {
  return strconv.Itoa(int(i)), nil
}
func (i Int) GetInt() (int, error) {
  return int(i), nil
}

type Bool bool
func (b Bool) isVariable() {}

type String string
func (s String) isVariable() {}
func (s String) GetString() (string, error) {
  return string(s), nil
}
func (s String) GetInt() (int, error) {
  return strconv.Atoi(string(s))
}

type Expression string
func (e Expression) isVariable() {}
func (e Expression) GetString() (string, error) {
  return string(e), nil
}
func (e Expression) GetInt() (int, error) {
  return 0, nil
}

type Container map[string]Variable
func (c Container) isVariable() {}
