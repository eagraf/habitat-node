package transitions

import (
	"fmt"

	"github.com/eagraf/habitat-node/entities"
)

type UpdateBacknetTransition struct {
	OldCommunity *entities.Community `json:"old_community" mapstructure:"old_community"`
	NewCommunity *entities.Community `json:"new_community" mapstructure:"new_community"`
}

func (ub UpdateBacknetTransition) Type() TransitionType {
	return UpdateBacknetTransitionType
}

func (ub UpdateBacknetTransition) Reduce(state *entities.State) (*entities.State, error) {
	newState := *state
	if _, ok := state.Communities[ub.OldCommunity.ID]; !ok {
		return nil, fmt.Errorf("no community with id %s in state", ub.OldCommunity.ID)
	}
	if ub.OldCommunity.ID != ub.NewCommunity.ID {
		return nil, fmt.Errorf("old and new community ids do not match %s, %s", ub.OldCommunity.ID, ub.NewCommunity.ID)
	}

	if ub.OldCommunity.Backnet.Type != ub.NewCommunity.Backnet.Type {
		return nil, fmt.Errorf("switching backnet implementations is not supported")
	}

	*newState.Communities[ub.OldCommunity.ID].Backnet = *ub.NewCommunity.Backnet

	return &newState, nil
}
