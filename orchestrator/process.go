package main

import (
	"context"

	"github.com/eagraf/habitat-node/entities"
)

type processType string

const (
	processTypeApp     processType = "app"
	processTypeBacknet processType = "backnet"
)

type processID string

type process struct {
	ID          processID
	communityID entities.CommunityID
	processType processType
	context     context.Context
	cancel      context.CancelFunc
	errChan     chan error
}
