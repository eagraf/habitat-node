package transitions

import (
	"fmt"

	"github.com/eagraf/habitat-node/entities"
)

type UpdateBacknetTransition struct {
	CommID     entities.CommunityID `json:"community_id" mapstructure:"community_id"`
	OldBacknet *entities.Backnet    `json:"old_backnet" mapstructure:"old_backnet"`
	NewBacknet *entities.Backnet    `json:"new_backnet" mapstructure:"new_backnet"`
}

func (ub UpdateBacknetTransition) Type() TransitionType {
	return UpdateBacknetTransitionType
}

func (ub UpdateBacknetTransition) CommunityID() entities.CommunityID {
	return ub.CommID
}

func (ub UpdateBacknetTransition) Reduce(oldCommunity *entities.Community) (*entities.Community, error) {
	newCommunity, err := oldCommunity.Copy()
	if err != nil {
		return nil, fmt.Errorf("error copying community: %s", err.Error())
	}

	if ub.OldBacknet.Type != ub.NewBacknet.Type {
		return nil, fmt.Errorf("switching backnet implementations is not supported")
	}

	newCommunity.Backnet = ub.NewBacknet

	return newCommunity, nil
}
