package app

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/eagraf/habitat-node/fslib"
	"github.com/rs/zerolog/log"
)

// TODO later: reserved port to query and get other ones?

type CLIConfig struct {
	fsapi   string // localhost:port for fs
	commapi string // localhost:port for community manager
}

// Run a simple command line interface
func RunCLI(fsapi string, commapi string) {

	fs := &fslib.FSLibConfig{
		FStype: "IPFS",
		FSapi:  fsapi,
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Welcome to habitat node!\n% ")
	for scanner.Scan() {
		cmd := scanner.Text()
		if len(cmd) == 0 {
			continue
		}

		tokens := strings.Split(cmd, " ")
		switch tokens[0] {
		case "exit":
			// TODO: graceful shutdown
			fmt.Println("Exiting!")
			return
		case "ipfs":
			// TODO: nicely print ipfs info
			/*
				if err != nil {
					panic(err)
				}
			*/
			log.Info().Msg("unimplemented")
		case "comm":
			// TODO: nicely print out the "current community"
			// later this will be for switching communities, adding users etc.
			CommRoute(tokens[1:])

		case "fs":
			FsRoute(*fs, tokens[1:])

		default:
			fmt.Printf("Command \"%s\" not recognized\n", tokens[0])
		}
		fmt.Print("% ")
	}
}

// Router creates a simple framework for making subcommands easy
type Router interface {
	Route(cmd []string)
}

func CommRoute(cmd []string) {
	if len(cmd) < 1 {
		log.Error().Err(errors.New("No subroute provided")).Msg("")
	}
	log.Debug().Str("command", cmd[0]).Msg("CommRoute, unimplemented")
}

func FsRoute(fs fslib.FSLibConfig, cmd []string) {
	if len(cmd) < 1 {
		log.Error().Err(errors.New("No subroute provided")).Msg("")
		return
	}
	log.Debug().Str("command", cmd[0]).Msg("FsRoute")

	switch cmd[0] {
	case "ls":
		if len(cmd) < 2 {
			log.Error().Err(errors.New("Not enough arguments provided")).Msg("")
			return
		}
		res, err := fs.Ls(cmd[1])
		if err != nil {
			log.Info().Str("LS Response", res)
		} else {
			log.Error().Err(err)
		}

	case "write":
		if len(cmd) < 3 {
			log.Error().Err(errors.New("Not enough arguments provided")).Msg("")
			return
		}
		res, err := fs.Write(cmd[1], cmd[2])
		if err != nil {
			log.Info().Str("Write Response", res)
		} else {
			log.Error().Err(err)
		}

	case "pin":
		if len(cmd) < 3 {
			log.Error().Err(errors.New("Not enough arguments provided")).Msg("")
			return
		}
		res, err := fs.Pin(cmd[1], cmd[2])
		if err != nil {
			log.Info().Str("Pin Response", res)
		} else {
			log.Error().Err(err)
		}

	case "remove":
		if len(cmd) < 2 {
			log.Error().Err(errors.New("Not enough arguments provided")).Msg("")
			return
		}
		res, err := fs.Remove(cmd[1])
		if err != nil {
			log.Info().Str("Remove Response", res)
		} else {
			log.Error().Err(err)
		}

	case "cat":
		if len(cmd) < 2 {
			log.Error().Err(errors.New("Not enough arguments provided")).Msg("")
			return
		}
		res, err := fs.Cat(cmd[1])
		if err != nil {
			log.Info().Str("Cat Response", res)
		} else {
			log.Error().Err(err)
		}

	case "move":
		if len(cmd) < 3 {
			log.Error().Err(errors.New("Not enough arguments provided")).Msg("")
			return
		}
		res, err := fs.Move(cmd[1], cmd[2])
		if err != nil {
			log.Info().Str("Move Response", res)
		} else {
			log.Error().Err(err)
		}

	case "copy":
		if len(cmd) < 3 {
			log.Error().Err(errors.New("Not enough arguments provided")).Msg("")
			return
		}
		res, err := fs.Copy(cmd[1], cmd[2])
		if err != nil {
			log.Info().Str("Copy Response", res)
		} else {
			log.Error().Err(err)
		}

	case "mkdir":
		if len(cmd) < 2 {
			log.Error().Err(errors.New("Not enough arguments provided")).Msg("")
			return
		}
		res, err := fs.Mkdir(cmd[1])
		if err != nil {
			log.Info().Str("Mkdir Response", res)
		} else {
			log.Error().Err(err)
		}

	default:
		log.Info().Msg("default case")
	}
}
