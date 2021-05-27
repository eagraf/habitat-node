package transitions

import (
	"testing"

	"github.com/eagraf/habitat-node/entities"
	"github.com/stretchr/testify/assert"
)

func TestAddCommunity(t *testing.T) {
	host := entities.InitHost()
	community := entities.InitCommunity("community_0", "My Community", entities.IPFS)

	transition := AddCommunityTransition{
		Community: community,
	}

	newHost, err := transition.Reduce(host)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(newHost.Communities))
}
