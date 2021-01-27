package client

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func main() {

	err := os.MkdirAll(os.Getenv("AUTH_DIR"), 0600)
	if err != nil {
		panic(err)
	}

	ur, err := NewUserRepo(os.Getenv("AUTH_DIR"))
	if err != nil {
		panic(err)
	}

	tr, err := NewTokenRepo(os.Getenv("AUTH_DIR"))
	if err != nil {
		panic(err)
	}

	as, err := NewAuthService(tr, ur)
	if err != nil {
		panic(err)
	}

	us := NewUserService(ur)
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter()

	// These routes don't require verification
	router.Path("/api/v1/login").Handler(http.HandlerFunc(as.LoginHandler))
	router.Path("/api/v1/bootstrap").Handler(http.HandlerFunc(us.BootstrapUserHandler))
	router.HandleFunc("/api/v1/version", VersionHandler)

	// api subrouter requires token authentication
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(as.Middleware)
	api.HandleFunc("/logout", as.LogoutHandler).Methods("POST")
	api.HandleFunc("/users", us.CreateUserHandler).Methods("POST")

	server := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:3000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("Client listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}

// VersionResponse contains basic version info about the group space
type VersionResponse struct {
	ClientVersion string `json:"client_version"`
}

// VersionHandler returns version information to user
func VersionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	version := VersionResponse{
		ClientVersion: "0.0.1",
	}
	body, err := json.Marshal(&version)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(body)
}
