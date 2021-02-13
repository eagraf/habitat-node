package main

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/eagraf/habitat-node/app"
	"github.com/eagraf/habitat-node/client"
	"github.com/eagraf/habitat-node/entities"
	"github.com/eagraf/habitat-node/fs"
	"github.com/rs/zerolog/log"
)

type processManager struct {
	processes map[processID]*Process
	backnets  map[entities.CommunityID]Backnet
	fs        ProcessID
	auth      ProcessID
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
		processes:  make(map[processID]*Process),
		backnets:   make(map[entities.CommunityID]Backnet),
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
		nets[community.ID] = *community.Backnet

		go func(community *entities.Community) {
			backnet, err := pm.startBacknet(community)
			if err != nil {
				log.Err(fmt.Errorf("error starting %s process for community %s: %s", community.Backnet.Type, community.ID, err.Error())).Msg("")
			}
			// clean this up later EGEGOEKWG
			if community.Backnet.Type == entities.IPFS {
				apiports[community.ID] = strconv.Itoa(backnet.(*IPFSBacknet).backnet.Local.PortMap["api"])
			}
		}(community)
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

func (pm *processManager) processErrorListener(process *Process) {
	for {
		err := <-process.errChan
		pm.errChan <- processError{
			processID:   process.ID,
			communityID: process.CommunityID,
			err:         err,
		}
	}
}

func (pm *processManager) startBacknet(community *entities.Community) (Backnet, error) {
	var backnet Backnet

	process := InitProcess(ProcessTypeBacknet)

	switch community.Backnet.Type {
	case entities.IPFS:
		myBacknet, err := InitIPFSBacknet(community, process)
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
		return nil, err
	}
	process, err = backnet.StartProcess()
	if err != nil {
		return nil, err
	}
	pm.processes[process.ID] = process
	pm.backnets[community.ID] = backnet

	go pm.processErrorListener(process)
	log.Info().Msgf("process %s started", process.ID)

	return backnet, nil
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
		_, err := pm.startBacknet(addCommunityTransition.Community)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("transition type %s not supported", transition.Type())
	}
	return nil
}
