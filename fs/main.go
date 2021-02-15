package fs

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/eagraf/habitat-node/client"
	"github.com/eagraf/habitat-node/entities"
	"github.com/gorilla/mux"
)

// RunFilesystem exported to be called by orchestrator
func RunFilesystem(as *client.AuthService, state *entities.State, ports map[entities.CommunityID]string, enets map[entities.CommunityID]entities.Backnet) {

	backnets := make(map[entities.CommunityID]Backnet)
	for id, api := range ports {
		if enet, ok := enets[id]; ok {
			splt := strings.Split(api, "/")
			port := splt[len(splt)-1]
			fmt.Print("127.0.0.1:" + port + "\n")
			backnets[id] = InitIPFSBacknet(id, enet, "127.0.0.1:"+port)
		}
	}

	fs, err := NewFilesystemService(as, state, backnets)
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter()
	router.PathPrefix("/api/fs/ls").Handler(http.HandlerFunc(fs.ParseListFiles))
	router.PathPrefix("/api/fs/write").Handler(http.HandlerFunc(fs.Write))

	// eventually want to do this:
	// api := router.PathPrefix("/api/v1/fs").Subrouter()
	// api.Use(fs.authService.Middleware)

	server := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:6000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("Filesystem API listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
