package main

import (
	"fmt"

	"github.com/eagraf/habitat-node/entities"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type processManager struct {
	processes map[processID]struct{}
	apps      map[processID]process
	backnets  map[processID]process
	errChan   chan processError
}

type processError struct {
	processID   processID
	communityID entities.CommunityID
	err         error
}

func initManager() *processManager {
	return &processManager{
		processes: make(map[processID]struct{}),
		apps:      make(map[processID]process),
		backnets:  make(map[processID]process),
		errChan:   make(chan processError),
	}
}

func (pm *processManager) start(state *entities.State) error {
	go pm.errorListener()
	for _, community := range state.Communities {
		var backnet Backnet
		switch community.Backnet.Type {
		case entities.IPFS:
			myBacknet := InitIPFSBacknet(&community)
			backnet = myBacknet
		case entities.DAT:
			fallthrough
		default:
			log.Err(fmt.Errorf("backnet type %s is not supported", community.Backnet.Type)).Msg("")
		}
		go func() {
			process, err := backnet.StartProcess()
			if err != nil {
				log.Err(fmt.Errorf("error starting %s process for community %s", community.Backnet.Type, community.ID))
			}
			process.ID = processID(uuid.New().String())
			pm.backnets[process.ID] = *process

			log.Info().Msgf("starting %s process for community %s %s", process.processType, process.communityID, process.ID)
			go pm.processErrorListener(process)
		}()
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
