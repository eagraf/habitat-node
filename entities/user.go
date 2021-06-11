package entities

import (
	"encoding/json"
	"errors"
)

// UserID identifies users. Uniqueness is maintained by a global namespace running on smart contracts in the Ether.
type UserID string

// User represents a person's account, which can belong to many communities.
type User struct {
	ID     UserID `json:"id"`
	Handle string `json:"handle"`

	Communities []CommunityID
}

func InitUser(id UserID, handle string) *User {
	return &User{
		ID:          id,
		Handle:      handle,
		Communities: make([]CommunityID, 0),
	}
}

func (u *User) Copy() (*User, error) {
	// dirty trick for copying: just marshal and unmarshal.
	// if performance is a huge issue, we can eventually create real copy methods

	marshalled, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}

	var copy User
	err = json.Unmarshal(marshalled, &copy)
	if err != nil {
		return nil, err
	}

	return &copy, err
}

func valInSlice(id CommunityID, comms []CommunityID) bool {
	for _, x := range comms {
		if id == x {
			return true
		}
	}
	return false
}

func removeValInSlice(id CommunityID, comms []CommunityID) []CommunityID {
	idx := -1
	for i, x := range comms {
		if id == x {
			idx = i
		}
	}
	if idx == -1 {
		return nil
	}
	comms[idx] = comms[len(comms)-1] // Copy last element to index i.
	comms[len(comms)-1] = ""         // Erase last element (write zero value).
	comms = comms[:len(comms)-1]     // Truncate slice.
	return comms
}

func (u *User) AddCommunity(id CommunityID) error {
	if valInSlice(id, u.Communities) {
		return errors.New("User is already part of the community")
	}
	u.Communities = append(u.Communities, id)
	return nil
}

func (u *User) RemoveCommunity(id CommunityID) error {
	if !valInSlice(id, u.Communities) {
		return errors.New("User is already not part of the community")
	}
	u.Communities = removeValInSlice(id, u.Communities)
	return nil
}
