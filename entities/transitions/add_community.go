package transitions

import (
	"fmt"

	"github.com/eagraf/habitat-node/entities"
)

type AddCommunityTransition struct {
	Community *entities.Community `json:"community"`
}

func (ac AddCommunityTransition) Type() TransitionType {
	return AddCommunityTransitionType
}

func (ac AddCommunityTransition) Reduce(oldHost *entities.Host) (*entities.Host, error) {
	newHost, err := oldHost.Copy()
	if err != nil {
		return nil, err
	}

	if _, ok := newHost.Communities[ac.Community.ID]; ok {
		return nil, fmt.Errorf("community with id %s is already in host", ac.Community.ID)
	}
	newHost.Communities[ac.Community.ID] = ac.Community
	return newHost, nil
}
