package state

import (
	"bytes"
	"testing"

	"github.com/eagraf/habitat-node/entities"
	"github.com/stretchr/testify/assert"
)

func TestSnapshot(t *testing.T) {
	myCommunity := &entities.Community{
		Name: "My Community",
	}

	buf := bytes.NewBuffer(make([]byte, 0))
	err := WriteSnapshot(buf, myCommunity, 42)
	assert.Nil(t, err)

	var newCommunity entities.Community
	sn, err := ReadSnapshot(buf, &newCommunity)
	assert.Nil(t, err)
	assert.Equal(t, uint64(42), sn.SequenceNumber)
	assert.Equal(t, "My Community", newCommunity.Name)
}
