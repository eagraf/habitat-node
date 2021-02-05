package main

import (
	"fmt"
	"sync"

	"github.com/eagraf/habitat-node/app"
	"github.com/eagraf/habitat-node/client"
	"github.com/eagraf/habitat-node/entities"
	"github.com/eagraf/habitat-node/fs"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type processManager struct {
	processes map[processID]struct{}
	apps      map[processID]process
	backnets  map[processID]process
	fs        processID
	auth      processID
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
		fs:         "",
		auth:       "",
		errChan:    make(chan processError),
		portMutex:  sync.Mutex{},
		portAllocs: make(map[int]processID),
		startPort:  4000,
		portCount:  0,
	}
}

func (pm *processManager) start(state *entities.State) error {
	go pm.errorListener()

	nets := make(map[entities.CommunityID]entities.Backnet)
	apiports := make(map[entities.CommunityID]string)

	for _, community := range state.Communities {
		nets[community.ID] = community.Backnet
		pid := processID(uuid.New().String())
		log.Info().Msgf("starting backnet process for community %s %s", community.ID, pid)
		var backnet Backnet
		switch community.Backnet.Type {
		case entities.IPFS:
			myBacknet, err := InitIPFSBacknet(&community)
			if err != nil {
				log.Err(err).Msg("error initializing backnet")
			}
			backnet = myBacknet
			apiports[community.ID] = myBacknet.config.Addresses.API[0]
			log.Info().Msgf("swarm port: %d, api port: %d, gateway port: %d", ports[0], ports[1], ports[2])
		case entities.DAT:
			fallthrough
		default:
			log.Err(fmt.Errorf("backnet type %s is not supported", community.Backnet.Type)).Msg("")
		}
		go func(pid processID, community entities.Community) {
			err := backnet.Configure(&community.Backnet)
			if err != nil {
				log.Err(fmt.Errorf("error configuring %s process for community %s: %s", community.Backnet.Type, community.ID, err.Error())).Msg("")
				return
			}
			process, err := backnet.StartProcess()
			if err != nil {
				log.Err(fmt.Errorf("error starting %s process for community %s: %s", community.Backnet.Type, community.ID, err.Error())).Msg("")
				return
			}
			process.ID = pid
			pm.backnets[process.ID] = *process
			pm.processes[process.ID] = struct{}{}

			go pm.processErrorListener(process)
			log.Info().Msgf("process %s started", process.ID)
		}(pid, community)
	}

	// add ports here?
	cli := client.InitClient()

	// is go the right way to kick off these processes?
	go cli.RunClient()
	go fs.RunFilesystem(cli.GetAuthService(), state, apiports, nets)
	go app.RunCLI("127.0.0.1:6000", "")

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
