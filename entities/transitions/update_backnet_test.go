package transitions

import (
	"testing"

	"github.com/eagraf/habitat-node/entities"
	"github.com/stretchr/testify/assert"
)

func TestUpdateBacknet(t *testing.T) {
	oldBacknet := &entities.Backnet{
		Bootstrap: []string{"bootstrap1", "bootstrap2"},
		Local: entities.LocalBacknetConfig{
			PortMap: map[string]int{"swarm": 4001, "api": 4002, "gateway": 4003},
		},
	}

	newBacknet := &entities.Backnet{
		Bootstrap: []string{"bootstrap1", "bootstrap2", "bootstrap3"},
		Local: entities.LocalBacknetConfig{
			PortMap: map[string]int{"swarm": 4004, "api": 4005, "gateway": 4006},
		},
	}

	oldCommunity := entities.Community{
		Backnet: oldBacknet,
	}

	transition := UpdateBacknetTransition{
		OldBacknet: oldBacknet,
		NewBacknet: newBacknet,
	}

	newCommunity, err := transition.Reduce(&oldCommunity)
	assert.Nil(t, err)

	newStateBacknet := newCommunity.Backnet
	assert.Equal(t, 3, len(newStateBacknet.Bootstrap))
	assert.Equal(t, 4005, newStateBacknet.Local.PortMap["api"])

	// Test attempt to change backnet type
	transition.NewBacknet.Type = entities.DAT
	_, err = transition.Reduce(&oldCommunity)
	assert.NotNil(t, err)
}
