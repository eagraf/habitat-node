package client

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

// Client is for other packages to access authService, add other data as needed
type Client struct {
	authService *AuthService
	userService *UserService
}

// InitClient just initializes empty services & directories
func InitClient() *Client {

	_, ok := os.LookupEnv("AUTH_DIR")
	if !ok {
		os.Setenv("AUTH_DIR", "auth/")
	}

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

	return &Client{
		authService: as,
		userService: us,
	}
}

// RunClient runs the client module (to be called by orchestrator)
func (client *Client) RunClient() {

	router := mux.NewRouter()

	// These routes don't require verification
	router.Path("/api/v1/login").Handler(http.HandlerFunc(client.authService.LoginHandler))
	router.Path("/api/v1/bootstrap").Handler(http.HandlerFunc(client.userService.BootstrapUserHandler))
	router.HandleFunc("/api/v1/version", VersionHandler)

	// api subrouter requires token authentication
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(client.authService.Middleware)
	api.HandleFunc("/logout", client.authService.LogoutHandler).Methods("POST")
	api.HandleFunc("/users", client.userService.CreateUserHandler).Methods("POST")

	server := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:3000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("Client listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())

}

// GetAuthService returns the corresponding authservice for use by other packages
func (cli *Client) GetAuthService() *AuthService {
	return cli.authService
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
