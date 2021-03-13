package main

import (
	"fmt"

	"github.com/eagraf/habitat-node/entities"
	"github.com/spf13/pflag"
)

type CriticalStateProcess struct {
	replicatedStateMachines map[entities.CommunityID]entities.ReplicatedStateMachine
}

func main() {

	// Flags to look for
	// - log location (log contains wal and snapshots)
	flags := pflag.NewFlagSet("state", pflag.ExitOnError)
	flags.StringP("logdir", "l", "", "specify a directory to store the log and snapshots")

	// Look for critical state stable storage

	// If nonexistent, initialize new state log

	csp := CriticalStateProcess{
		replicatedStateMachines: make(map[entities.CommunityID]entities.ReplicatedStateMachine),
	}

	var state *entities.State

	for _, community := range state.Communities {
		rsm, err := initRSM(&community)
		if err != nil {
			fmt.Println("AAHHH")
		}
		csp.replicatedStateMachines[community.ID] = rsm
		go rsm.Start() // TODO some way of keeping track of rsm errors
	}

}

func initRSM(community *entities.Community) (entities.ReplicatedStateMachine, error) {
	switch community.ConsensusAlgorithm.Type {
	case entities.Raft:
		return nil, nil
	default:
		return nil, fmt.Errorf("no implementation for consensus algorithm %s available", community.ConsensusAlgorithm.Type)
	}
}
