package main

import (
	"errors"
	"fmt"
	"sync"

	"github.com/eagraf/habitat-node/entities"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type processManager struct {
	processes map[processID]struct{}
	apps      map[processID]process
	backnets  map[processID]process
	errChan   chan processError

	portMutex  sync.Mutex
	portAllocs map[int]processID
	startPort  int
	portCount  int
}

type processError struct {
	processID   processID
	communityID entities.CommunityID
	err         error
}

func initManager() *processManager {
	return &processManager{
		processes:  make(map[processID]struct{}),
		apps:       make(map[processID]process),
		backnets:   make(map[processID]process),
		errChan:    make(chan processError),
		portMutex:  sync.Mutex{},
		portAllocs: make(map[int]processID),
		startPort:  4000,
		portCount:  0,
	}
}

func (pm *processManager) start(state *entities.State) error {
	go pm.errorListener()
	for _, community := range state.Communities {
		go func(community *entities.Community) {
			err := pm.startBacknet(community)
			if err != nil {
				log.Err(fmt.Errorf("error starting %s process for community %s: %s", community.Backnet.Type, community.ID, err.Error())).Msg("")
			}
		}(community)
	}

	return nil
}

func (pm *processManager) errorListener() {
	for {
		pErr := <-pm.errChan
		log.Err(pErr.err).Msgf("process: %s, community: %s", pErr.processID, pErr.communityID)
	}
}

func (pm *processManager) processErrorListener(process *process) {
	for {
		err := <-process.errChan
		pm.errChan <- processError{
			processID:   process.ID,
			communityID: process.communityID,
			err:         err,
		}
	}
}

func (pm *processManager) startBacknet(community *entities.Community) error {
	var backnet Backnet
	pid := processID(uuid.New().String())

	switch community.Backnet.Type {
	case entities.IPFS:
		myBacknet, err := InitIPFSBacknet(community)
		if err != nil {
			log.Err(err).Msg("error initializing backnet")
		}
		backnet = myBacknet
	case entities.DAT:
		fallthrough
	default:
		log.Err(fmt.Errorf("backnet type %s is not supported", community.Backnet.Type)).Msg("")
	}

	err := backnet.Configure(community.Backnet)
	if err != nil {
		return err
	}
	process, err := backnet.StartProcess()
	if err != nil {
		return err
	}
	process.ID = pid
	pm.backnets[process.ID] = *process
	pm.processes[process.ID] = struct{}{}

	go pm.processErrorListener(process)
	log.Info().Msgf("process %s started", process.ID)

	return nil
}

// Receive implements TransitionSubscriber
func (pm *processManager) Receive(transition entities.Transition) error {
	switch transition.Type() {
	case entities.AddCommunityTransitionType:
		addCommunityTransition, ok := transition.(entities.AddCommunityTransition)
		if !ok {
			return errors.New("transition is not type AddCommunityTransition")
		}

		// start communities backnet
		err := pm.startBacknet(addCommunityTransition.Community)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("transition type %s not supported", transition.Type())
	}
	return nil
}
