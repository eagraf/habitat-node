package transitions

import (
	"testing"

	"github.com/eagraf/habitat-node/entities"
	"github.com/stretchr/testify/assert"
)

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
