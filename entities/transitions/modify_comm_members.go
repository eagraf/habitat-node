package transitions

import (
	"errors"

	"github.com/eagraf/habitat-node/entities"
)

type ModifyType string

// All possible TransitionTypes
const (
	AddMember    ModifyType = "ADD_MEMBER"
	RemoveMember ModifyType = "REMOVE_MEMBER"
	BanMember    ModifyType = "BAN_MEMBER"
)

type ModifyCommMembersTransition struct {
	Community *entities.Community `json:"community"`
	User      *entities.User      `json:"user"`
	ModType   ModifyType          `json:"type"`
}

func (mt ModifyCommMembersTransition) Type() TransitionType {
	return ModifyCommMembersTransitionType
}

func (mt ModifyCommMembersTransition) CommunityID() entities.CommunityID {
	return mt.Community.ID
}

func (mt ModifyCommMembersTransition) Reduce(oldComm *entities.Community) (*entities.Community, error) {
	newCommunity, err := mt.Community.Copy()

	if err != nil {
		return nil, err
	}

	switch mt.ModType {
	case AddMember:
		err = newCommunity.AddMember(mt.User)
	case RemoveMember:
		err = newCommunity.RemoveMember(mt.User)
	case BanMember:
		err = errors.New("Banning members from communities is unimplemented!")
	}

	if err != nil {
		return nil, err
	}
	return newCommunity, err
}
