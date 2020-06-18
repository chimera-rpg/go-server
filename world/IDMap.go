package world

type ID = int32

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
