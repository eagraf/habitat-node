package state

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/eagraf/habitat-node/entities"
	"github.com/stretchr/testify/assert"
)

func TestWriteAhead(t *testing.T) {
	writer := bytes.Buffer{}
	log := &Log{
		LogWriter:         &writer,
		CurSequenceNumber: 0,
	}

	transition1 := &entities.TransitionWrapper{
		Type: entities.AddCommunityTransitionType,
		Transition: entities.AddCommunityTransition{
			Community: &entities.Community{
				Name: "my_community",
			},
		},
	}

	transition2 := &entities.TransitionWrapper{
		Type: entities.UpdateBacknetTransitionType,
		Transition: entities.UpdateBacknetTransition{
			OldCommunity: &entities.Community{},
			NewCommunity: &entities.Community{},
		},
	}

	transition3 := &entities.TransitionWrapper{
		Type: entities.AddCommunityTransitionType,
		Transition: entities.AddCommunityTransition{
			Community: &entities.Community{
				Name: "community_2",
			},
		},
	}

	err := log.WriteAhead(transition1)
	assert.Nil(t, nil, err)
	res := writer.String()
	n := strings.Count(res, "\n")
	assert.Equal(t, 1, n)

	err = log.WriteAhead(transition2)
	assert.Nil(t, nil, err)
	res = writer.String()
	n = strings.Count(res, "\n")
	assert.Equal(t, 2, n)

	err = log.WriteAhead(transition3)
	assert.Nil(t, nil, err)
	res = writer.String()
	n = strings.Count(res, "\n")
	assert.Equal(t, 3, n)

	// try decoding all of the entries
	entries := strings.Split(res, "\n")
	for i, entry := range entries {
		if len(entry) != 0 {
			fmt.Println(len(entry))
			decoded, err := DecodeLogEntry([]byte(entry))
			assert.Nil(t, err)
			assert.Equal(t, int64(i), decoded.SequenceNumber)
		}
	}
}
