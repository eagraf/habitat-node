package transitions

import (
	"testing"

	"github.com/eagraf/habitat-node/entities"
	"github.com/stretchr/testify/assert"
)

func TestUpdateBacknet(t *testing.T) {
	state := entities.InitState()
	oldCommunity := entities.InitCommunity("community_0", "My Community", entities.IPFS)
	oldCommunity.Backnet.Bootstrap = []string{"boostrap1", "bootstrap2"}
	oldCommunity.Backnet.Local.PortMap = map[string]int{"swarm": 4001, "api": 4002, "gateway": 4003}
	state.Communities["community_0"] = oldCommunity

	newCommunity := entities.InitCommunity("community_0", "My Community", entities.IPFS)
	newCommunity.Backnet.Bootstrap = []string{"boostrap1", "bootstrap2", "bootstrap3"}
	newCommunity.Backnet.Local.PortMap = map[string]int{"swarm": 4004, "api": 4005, "gateway": 4006}

	transition := UpdateBacknetTransition{
		OldCommunity: oldCommunity,
		NewCommunity: newCommunity,
	}

	newState, err := transition.Reduce(state)
	assert.Nil(t, err)

	newStateBacknet := newState.Communities["community_0"].Backnet
	assert.Equal(t, 3, len(newStateBacknet.Bootstrap))
	assert.Equal(t, 4005, newStateBacknet.Local.PortMap["api"])

	// test communities with different ids
	newCommunity.ID = "bad"
	_, err = transition.Reduce(state)
	assert.NotNil(t, err)
	newCommunity.ID = "community_0"

	// Test attempt to change backnet type
	newCommunity.Backnet.Type = entities.DAT
	_, err = transition.Reduce(state)
	assert.NotNil(t, err)
}
