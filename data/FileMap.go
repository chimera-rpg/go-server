package data

import (
	"errors"
	"fmt"
	"hash/crc32"
	"io/ioutil"
)

var fileMapTable = crc32.MakeTable(crc32.Koopman)

// FileID represents unique file string IDs
type FileID = uint32

// FileMap provides a centralized location for referring to
// filepaths via ids, and holding filepath data to CRC32 checksums.
// These are built upon server load.
type FileMap struct {
	Paths     map[FileID]string
	Checksums map[FileID]uint32
}

// BuildCRC builds a data CRC for a given id with a provided filepath
func (f *FileMap) BuildCRC(id FileID, filepath string) (uint32, error) {
	if val, ok := f.Checksums[id]; ok {
		if val != 0 {
			return val, nil
		}
	}

	r, err := ioutil.ReadFile(filepath)
	if err != nil {
		return 0, err
	}

	f.Paths[id] = filepath
	f.Checksums[id] = crc32.Checksum(r, fileMapTable)

	return f.Checksums[id], nil
}

// GetBytes returns the bytes corresponding to a FileID.
func (f *FileMap) GetBytes(id FileID) (bytes []byte, err error) {
	if p, ok := f.Paths[id]; ok {
		return ioutil.ReadFile(p)
	}
	err = errors.New(fmt.Sprintf("non-existent file id \"%d\" requested", id))
	return
}

// NewFileMap returns a constructed FileMap.
func NewFileMap() FileMap {
	return FileMap{
		Paths:     make(map[FileID]string),
		Checksums: make(map[FileID]uint32),
	}
}
