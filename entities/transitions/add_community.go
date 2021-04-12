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

func (ac AddCommunityTransition) Reduce(state *entities.State) (*entities.State, error) {
	newState := *state
	if _, ok := state.Communities[ac.Community.ID]; ok {
		return nil, fmt.Errorf("community with id %s is already in state", ac.Community.ID)
	}
	newState.Communities[ac.Community.ID] = ac.Community
	return &newState, nil
}

func (ac AddCommunityTransition) CommunityID() entities.CommunityID {
	return ac.Community.ID
}
