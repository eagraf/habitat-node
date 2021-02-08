package fs

import (
	"net/http"

	"github.com/eagraf/habitat-node/client"
	"github.com/eagraf/habitat-node/entities"
	"github.com/gorilla/mux"
)

// RunFilesystem exported to be called by orchestrator
func RunFilesystem(as *client.AuthService, state *entities.State, ports map[entities.CommunityID]string, enets map[entities.CommunityID]entities.Backnet) {

	backnets := make(map[entities.CommunityID]Backnet)
	for id, port := range ports {
		if enet, ok := enets[id]; ok {
			backnets[id] = InitIPFSBacknet(id, enet, port)
		}
	}

	fs, err := NewFilesystemService(as, state, backnets)
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter()
	router.Path("api/v1/app/ls").Handler(http.HandlerFunc(fs.ParseListFiles))

}
