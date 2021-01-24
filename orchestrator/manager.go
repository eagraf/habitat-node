package main

import (
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
		pid := processID(uuid.New().String())
		log.Info().Msgf("starting backnet process for community %s %s", community.ID, pid)
		var backnet Backnet
		switch community.Backnet.Type {
		case entities.IPFS:
			ports := pm.allocatePorts(pid, 3)
			myBacknet, err := InitIPFSBacknet(&community, ports[0], ports[1], ports[2])
			if err != nil {
				log.Err(err).Msg("error initializing backnet")
			}
			backnet = myBacknet
			log.Info().Msgf("swarm port: %d, api port: %d, gateway port: %d", ports[0], ports[1], ports[2])
		case entities.DAT:
			fallthrough
		default:
			log.Err(fmt.Errorf("backnet type %s is not supported", community.Backnet.Type)).Msg("")
		}
		go func(pid processID, community entities.Community) {
			process, err := backnet.StartProcess()
			if err != nil {
				log.Err(fmt.Errorf("error starting %s process for community %s", community.Backnet.Type, community.ID))
				return
			}
			process.ID = pid
			pm.backnets[process.ID] = *process
			pm.processes[process.ID] = struct{}{}

			go pm.processErrorListener(process)
			log.Info().Msgf("process %s started", process.ID)
		}(pid, community)
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

func (pm *processManager) allocatePorts(processID processID, n int) []int {
	pm.portMutex.Lock()
	defer pm.portMutex.Unlock()

	ports := make([]int, n, n)
	for i := 0; i < n; i++ {
		port := pm.startPort + pm.portCount + i
		pm.portAllocs[port] = processID
		ports[i] = port
	}
	pm.portCount += n

	return ports
}
