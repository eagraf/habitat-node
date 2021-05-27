package transitions

import (
	"fmt"

	"github.com/eagraf/habitat-node/entities"
)

type AddMemberTransition struct {
	Community entities.CommunityID `json:"community_id"`
	User      *entities.User       `json:"user`
}

func (am AddMemberTransition) Type() TransitionType {
	return AddMemberTransitionType
}

func (am AddMemberTransition) Reduce(state *entities.State) (*entities.State, error) {
	newState := *state
	if _, ok := state.Communities[am.Community]; !ok {
		return nil, fmt.Errorf("no community with id %s in state", am.Community)
	}
	if _, ok := state.Communities[am.Community].Members[am.User.ID]; ok {
		return nil, fmt.Errorf("community %s already has member %s", am.Community, am.User)
	}
	newState.Communities[am.Community].Members[am.User.ID] = am.User
	return &newState, nil
}

func (am AddMemberTransition) CommunityID() entities.CommunityID {
	return am.Community
}
