package entities

// CommunityID identifies a Community
type CommunityID string

// A Community is a collection of users, associated with a backnet and a collection of apps
type Community struct {
	ID      CommunityID `json:"id"`
	Name    string      `json:"name"`
	Members []UserID    `json:"members"`
	Peers   []Peer      `json:"peers"`
}
