package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/eagraf/habitat-node/entities"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Info().Msg("orchestrator starting")

	backnet := entities.InitBacknet(entities.IPFS)
	process := IPFSBacknet{
		backnet: *backnet,
	}

	_, err := process.StartProcess("community_0")
	if err != nil {
		log.Panic().Msg(err.Error())
	}
}
