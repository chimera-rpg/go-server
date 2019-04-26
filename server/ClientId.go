package server

// acquireClientId gets an unused ClientId(int) from either the slice of
// unusedClientIDs or from topClientID if there are no recycled ClientId(s).
func (server *GameServer) acquireClientID() int {
	var id int
	if len(server.unusedClientIDs) > 0 {
		// Pop first free id and resize unusedClientIDs slice
		id, server.unusedClientIDs = server.unusedClientIDs[0], server.unusedClientIDs[1:]
	} else {
		server.topClientID++
		id = server.topClientID
	}
	return id
}

// releaseClientId places the given integer into the unusedClientIDs slice
func (server *GameServer) releaseClientID(id int) {
	server.unusedClientIDs = append(server.unusedClientIDs, id)
}
