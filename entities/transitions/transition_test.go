package transitions

import (
	"encoding/json"
	"testing"

	"github.com/eagraf/habitat-node/entities"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalTransitionWrapper(t *testing.T) {
	jsonBytes := []byte(`
{
	"type": "ADD_COMMUNITY",
	"transition": {
		"community": {
			"id": "community_0",
			"name": "my community"
		}
	}
}
	`)

	var tw TransitionWrapper
	err := json.Unmarshal(jsonBytes, &tw)
	assert.Nil(t, err)

	act, ok := tw.Transition.(*AddCommunityTransition)
	assert.Equal(t, true, ok)
	assert.Equal(t, "community_0", string(act.CommunityID()))
}

func TestAddCommunity(t *testing.T) {
	state := entities.InitState()
	community := entities.InitCommunity("community_0", "My Community", entities.IPFS)

	transition := AddCommunityTransition{
		Community: community,
	}

	_, err := transition.Reduce(state)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(state.Communities))
}

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
