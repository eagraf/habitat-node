package entities

// AppID is unique to any application, guaranteed by a global namespace on the Ether.
type AppID string

// App represents an application running on a community's servers.
type App struct {
	AppID   string `json:"id"`
	Running bool   `json:"running"`
}
