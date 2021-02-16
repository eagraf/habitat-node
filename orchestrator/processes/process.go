package processes

import (
	"context"

	"github.com/eagraf/habitat-node/entities"
	"github.com/google/uuid"
)

type ProcessType string

// Process types
const (
	ProcessTypeApp     ProcessType = "app"
	ProcessTypeBacknet ProcessType = "backnet"
)

type ProcessID string

type Process struct {
	ID          ProcessID
	CommunityID entities.CommunityID
	ProcessType ProcessType

	context context.Context
	cancel  context.CancelFunc
	errChan chan error
}

func InitProcess(pType ProcessType) *Process {
	return &Process{
		ID:          ProcessID(uuid.New().String()),
		ProcessType: pType,
	}
}
