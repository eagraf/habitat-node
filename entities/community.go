package entities

import "encoding/json"

// CommunityID identifies a Community
type CommunityID string

// A Community is a collection of users, associated with a backnet and a collection of apps
type Community struct {
	ID                 CommunityID               `json:"id"`
	Name               string                    `json:"name"`
	Members            map[UserID]*User          `json:"members"`
	Peers              []*Peer                   `json:"peers"`
	Backnet            *Backnet                  `json:"backnet"`
	Apps               []*AppID                  `json:"apps"`
	ConsensusAlgorithm *ConsensusAlgorithmConfig `json:"consensus_algorithm"`
}

func InitCommunity(id CommunityID, name string, backnetType BacknetType) *Community {
	return &Community{
		ID:      id,
		Name:    name,
		Members: make(map[UserID]*User),
		Peers:   make([]*Peer, 0),
		Backnet: InitBacknet(backnetType),
		Apps:    make([]*AppID, 0),
	}
}

func (c *Community) Copy() (*Community, error) {
	// dirty trick for copying: just marshal and unmarshal.
	// if performance is a huge issue, we can eventually create real copy methods

	marshalled, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	var copy Community
	err = json.Unmarshal(marshalled, &copy)
	if err != nil {
		return nil, err
	}

	return &copy, err
}
