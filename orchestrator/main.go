package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/eagraf/habitat-node/entities"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Info().Msg("orchestrator starting")

	state := &entities.State{
		Communities: map[entities.CommunityID]entities.Community{
			entities.CommunityID("community_0"): {
				ID: entities.CommunityID("community_0"),
				Backnet: entities.Backnet{
					Type: entities.IPFS,
				},
			},
		},
	}

	m := initManager()
	m.start(state)

	select {}
}
