package main

import (
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/eagraf/habitat-node/entities"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Info().Msg("orchestrator starting")

	state := &entities.State{
		Communities: map[entities.CommunityID]*entities.Community{
			entities.CommunityID("community_0"): {
				ID: entities.CommunityID("community_0"),
				Backnet: &entities.Backnet{
					Type: entities.IPFS,
					Local: entities.LocalBacknetConfig{
						PortMap: map[string]int{
							"swarm":   4001,
							"api":     4002,
							"gateway": 4003,
						},
					},
				},
			},
			entities.CommunityID("community_1"): {
				ID: entities.CommunityID("community_1"),
				Backnet: &entities.Backnet{
					Type: entities.IPFS,
					Local: entities.LocalBacknetConfig{
						PortMap: map[string]int{
							"swarm":   4004,
							"api":     4005,
							"gateway": 4006,
						},
					},
				},
			},
		},
	}

	m := initManager()
	m.start(state)

	time.Sleep(10 * time.Second)

	m.Receive(entities.AddCommunityTransition{
		Community: &entities.Community{
			ID: entities.CommunityID("community_2"),
			Backnet: &entities.Backnet{
				Type: entities.IPFS,
				Local: entities.LocalBacknetConfig{
					PortMap: map[string]int{
						"swarm":   4007,
						"api":     4008,
						"gateway": 4009,
					},
				},
			},
		},
	})

	select {}
}
