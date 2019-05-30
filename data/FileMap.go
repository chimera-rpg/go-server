package data

import (
	"hash/crc32"
	"io/ioutil"
)

var fileMapTable = crc32.MakeTable(crc32.Koopman)

type FileId = uint32

// DataMapping provides a centralized location for referring to
// filepaths via ids, and holding filepath data to CRC32 checksums.
// These are built upon server load.
type FileMap struct {
	Paths     map[FileId]string
	Checksums map[FileId]uint32
}

// BuildCRC builds a data CRC for a given id with a provided filepath
func (f *FileMap) BuildCRC(id FileId, filepath string) (uint32, error) {
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

func NewFileMap() FileMap {
	return FileMap{
		Paths:     make(map[FileId]string),
		Checksums: make(map[FileId]uint32),
	}
}
