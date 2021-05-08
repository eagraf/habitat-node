package transitions

import (
	"errors"

	"github.com/eagraf/habitat-node/entities"
)

type InitCommunityTransition struct {
	Community *entities.Community `json:"community"`
}

func (ic InitCommunityTransition) Type() TransitionType {
	return InitCommunityTransitionType
}

func (ic InitCommunityTransition) CommunityID() entities.CommunityID {
	return ic.Community.ID
}

func (ic InitCommunityTransition) Reduce(oldCommunity *entities.Community) (*entities.Community, error) {
	// sanity check, this should always be the first transition that sets initial state for the community
	// a not nil oldCommunity indicates a good change of something fishy going on
	if oldCommunity != nil {
		return nil, errors.New("oldCommunity is not nil")
	}

	newCommunity, err := ic.Community.Copy()
	if err != nil {
		return nil, err
	}

	return newCommunity, nil
}
