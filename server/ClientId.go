package server

// acquireClientId gets an unused ClientId(int) from either the slice of
// unusedClientIds or from topClientId if there are no recycled ClientId(s).
func (server *GameServer) acquireClientId() int {
	var id int
	if len(server.unusedClientIds) > 0 {
		// Pop first free id and resize unusedClientIds slice
		id, server.unusedClientIds = server.unusedClientIds[0], server.unusedClientIds[1:]
	} else {
		server.topClientId++
		id = server.topClientId
	}
	return id
}

// releaseClientId places the given integer into the unusedClientIds slice
func (server *GameServer) releaseClientId(id int) {
	server.unusedClientIds = append(server.unusedClientIds, id)
}
