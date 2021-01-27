package fs

import (
	"net/http"

	"github.com/eagraf/habitat-node/client"
	"github.com/eagraf/habitat-node/entities"
	"github.com/gorilla/mux"
)

// TODO: i shouldn't be creating a new user repo/token repo/auth service ...
// need to somehow connect them with client but also keep seprate...

func runFilesystemAPI(as *client.AuthService, state *entities.State, nets map[entities.CommunityID]Backnet) {

	fs, err := NewFilesystemService(as, state, nets)
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter()
	router.Path("api/v1/app/ls").Handler(http.HandlerFunc(fs.ParseListFiles))

}
