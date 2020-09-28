package world

// ID is a simple identifier.
type ID = uint32

// IDMap is a structure for managing a pool of IDs that can be used and freed.
type IDMap struct {
	id      ID
	freeIDs []ID
	usedIDs []ID
}

func (idm *IDMap) acquire() (id ID) {
	if len(idm.freeIDs) > 0 {
		id = idm.freeIDs[0]
		idm.freeIDs = append(idm.freeIDs[:0], idm.freeIDs[1:]...)
		return
	}
	idm.id++
	id = idm.id
	return
}

func (idm *IDMap) free(id ID) {
	idm.freeIDs = append(idm.freeIDs, id)
}
