package entities

// BacknetType enumerates types of backnets
type BacknetType string

// Available backnets that fulfill the Backnet interface
const (
	Local BacknetType = "local"
	IPFS  BacknetType = "ipfs"
	DAT   BacknetType = "dat"
)

// Backnet is a "backing network", which stores files (usually in a p2p filesystem)
type Backnet struct {
	Type      BacknetType `json:"type"`
	Bootstrap []string    `json:"bootstrap"`
}
