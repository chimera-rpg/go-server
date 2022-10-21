package data

import (
	"hash/crc32"
)

var stringMapTable = crc32.MakeTable(crc32.Koopman)

// StringID is a unique ID for a particular string
type StringID = uint32

// Strings provides a StringID to string map and reverse map.
type Strings struct {
	IDs     map[StringID]string
	Strings map[string]StringID
}

// Acquire returns the StringID that the provided name string corresponds to.
func (n *Strings) Acquire(name string) StringID {
	if val, ok := n.Strings[name]; ok {
		return val
	}
	id := crc32.Checksum([]byte(name), stringMapTable)

	n.IDs[id] = name
	n.Strings[name] = id

	return id
}

// Lookup reutrns the string that the provided StringID corresponds to.
func (n *Strings) Lookup(id StringID) string {
	if val, ok := n.IDs[id]; ok {
		return val
	}
	return ""
}

// NewStrings provides a constructed instance of Strings.
func NewStrings() Strings {
	return Strings{
		IDs:     make(map[StringID]string),
		Strings: make(map[string]StringID),
	}
}

// Whatever, let's have a global strings map.
var StringsMap = NewStrings()
