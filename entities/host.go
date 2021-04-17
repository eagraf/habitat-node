package entities

import "encoding/json"

// Host is the top level state object
type Host struct {
	Communities map[CommunityID]*Community `json:"communities"` // Communities that the node is helping to host
	HostUsers   map[string]HostUser        `json:"users"`       // Users that are logged into this node
}

func InitHost() *Host {
	return &Host{
		Communities: make(map[CommunityID]*Community),
		HostUsers:   make(map[string]HostUser),
	}
}

func (h *Host) Copy() (*Host, error) {
	// TODO using json encode/decode for copying is a hack, and is way less efficient than this could be
	marshalled, err := json.Marshal(h)
	if err != nil {
		return nil, err
	}

	var copy Host
	err = json.Unmarshal(marshalled, &copy)
	if err != nil {
		return nil, err
	}

	return &copy, nil
}

// HostUser represents an account that is allowed to configure the physical node
type HostUser struct {
	Username string
}
