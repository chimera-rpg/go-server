package data

import (
	"hash/crc32"
)

var stringMapTable = crc32.MakeTable(crc32.Koopman)

type StringId = uint32

type StringsMap struct {
	Ids     map[StringId]string
	Strings map[string]StringId
}

func (n *StringsMap) Acquire(name string) StringId {
	if val, ok := n.Strings[name]; ok {
		return val
	}
	id := crc32.Checksum([]byte(name), stringMapTable)

	n.Ids[id] = name
	n.Strings[name] = id

	return id
}

func NewStringsMap() StringsMap {
	return StringsMap{
		Ids:     make(map[StringId]string),
		Strings: make(map[string]StringId),
	}
}
