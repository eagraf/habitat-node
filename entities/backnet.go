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
	Type      BacknetType        `json:"type"`
	Bootstrap []string           `json:"bootstrap"`
	Local     LocalBacknetConfig `json:"local_backnet_config"`
}

// LocalBacknetConfig contains configurations for a backnet that are not shared with peers
type LocalBacknetConfig struct {
	PortMap map[string]int
}

func InitBacknet(backnetType BacknetType) *Backnet {
	return &Backnet{
		Type:      backnetType,
		Bootstrap: make([]string, 0),
	}
}
