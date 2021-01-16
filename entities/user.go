package entities

// UserID identifies users. Uniqueness is maintained by a global namespace running on smart contracts in the Ether.
type UserID string

// User represents a person's account, which can belong to many communities.
type User struct {
	ID     UserID `json:"id"`
	Handle string `json:"handle"`

	Communities []CommunityID
}
