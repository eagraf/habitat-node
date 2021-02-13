package main

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

type processID string

type Process struct {
	ID          processID
	CommunityID entities.CommunityID
	ProcessType ProcessType

	context context.Context
	cancel  context.CancelFunc
	errChan chan error
}

func InitProcess(pType ProcessType) *Process {
	return &Process{
		ID:          processID(uuid.New().String()),
		ProcessType: pType,
	}
}
