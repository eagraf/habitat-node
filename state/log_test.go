package state

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/eagraf/habitat-node/entities"
	"github.com/eagraf/habitat-node/entities/transitions"
	"github.com/stretchr/testify/assert"
)

func TestWriteAhead(t *testing.T) {
	writer := bytes.Buffer{}
	log := &Log{
		logWriter: &writer,
		mutex:     &sync.Mutex{},
	}

	transition1 := &transitions.TransitionWrapper{
		Type: transitions.AddCommunityTransitionType,
		Transition: transitions.AddCommunityTransition{
			Community: &entities.Community{
				Name: "my_community",
			},
		},
		SequenceNumber: 1,
	}

	transition2 := &transitions.TransitionWrapper{
		Type: transitions.UpdateBacknetTransitionType,
		Transition: transitions.UpdateBacknetTransition{
			OldBacknet: &entities.Backnet{},
			NewBacknet: &entities.Backnet{},
		},
		SequenceNumber: 2,
	}

	transition3 := &transitions.TransitionWrapper{
		Type: transitions.AddCommunityTransitionType,
		Transition: transitions.AddCommunityTransition{
			Community: &entities.Community{
				Name: "community_2",
			},
		},
		SequenceNumber: 3,
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
			decoded, err := DecodeLogEntry([]byte(entry))
			assert.Nil(t, err)
			assert.Equal(t, uint64(i+1), decoded.SequenceNumber)
		}
	}
}
