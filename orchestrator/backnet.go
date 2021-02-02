package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/eagraf/habitat-node/entities"
	"golang.org/x/net/context"
)

type Backnet interface {
	Configure(backnet *entities.Backnet) error
	StartProcess() (*process, error)
}

type IPFSBacknet struct {
	communityID entities.CommunityID
	backnet     entities.Backnet
	process     process

	ipfsDir    string
	configPath string
	config     *IPFSConfig
}

func InitIPFSBacknet(community *entities.Community, swarmPort, apiPort, gatewayPort int) (*IPFSBacknet, error) {
	// Make ipfs dir
	ipfsDir := filepath.Join(os.Getenv("IPFS_DIR"), string(community.ID))
	err := os.MkdirAll(ipfsDir, 0700)
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(os.Getenv("CONFIG_DIR"), string(community.ID))
	err = os.MkdirAll(configDir, 0700)
	if err != nil {
		return nil, err
	}

	builder, err := NewIPFSConfigBuilder()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "ipfs_config.json")
	builder.SetAddresses(swarmPort, apiPort, gatewayPort)
	config := builder.Config()
	config.WriteConfig(configPath)

	return &IPFSBacknet{
		communityID: community.ID,
		backnet:     community.Backnet,
		ipfsDir:     ipfsDir,
		configPath:  configPath,
		config:      config,
	}, nil
}

func (ib *IPFSBacknet) Configure(newBacknet *entities.Backnet) error {
	if newBacknet.Type != entities.IPFS {
		return errors.New("backnet should be of type IPFS")
	}

	// TODO validate LocalBacknetConfig

	ipfsConfigPath := filepath.Join(os.Getenv("IPFS_DIR"), string(ib.communityID), "config")

	// If the backnet has not been initialized before, run ipfs init
	if _, err := os.Stat(ipfsConfigPath); os.IsNotExist(err) {
		ipfsDir := filepath.Join(os.Getenv("IPFS_DIR"), string(ib.communityID))
		err := os.MkdirAll(ipfsDir, 0700)
		if err != nil {
			return err
		}

		configDir := filepath.Join(os.Getenv("CONFIG_DIR"), string(ib.communityID))
		err = os.MkdirAll(configDir, 0700)
		if err != nil {
			return err
		}

		// TODO factor out configuration logic and unit test

		builder, err := NewIPFSConfigBuilder()
		if err != nil {
			return err
		}

		swarmPort, ok := newBacknet.Local.PortMap["swarm"]
		if !ok {
			return errors.New("no swarm port included in port map")
		}
		apiPort, ok := newBacknet.Local.PortMap["api"]
		if !ok {
			return errors.New("no api port included in port map")
		}
		gatewayPort, ok := newBacknet.Local.PortMap["gateway"]
		if !ok {
			return errors.New("no gateway port included in port map")
		}

		configPath := filepath.Join(configDir, "ipfs_config.json")
		builder.SetAddresses(swarmPort, apiPort, gatewayPort)
		config := builder.Config()
		err = config.WriteConfig(configPath)
		if err != nil {
			return err
		}

		env := []string{
			fmt.Sprintf("IPFS_PATH=%s", ib.ipfsDir),
		}

		// Run ipfs init
		ctx, _ := context.WithCancel(context.Background())
		cmd := exec.CommandContext(ctx, "ipfs", "init", ib.configPath)
		cmd.Env = env
		//cmd.Stdout = os.Stdout
		//cmd.Stderr = os.Stderr
		err = cmd.Start()
		if err != nil {
			return err
		}
		err = cmd.Wait()
		if err != nil {
			return err
		}

		return nil

	} else {
		return errors.New("in flight reconfiguration not supported yet")
	}

	// Otherwise, make request to change config
	// TODO make sure daemon is restarted as well
}

func (ib *IPFSBacknet) StartProcess() (*process, error) {
	if ib.backnet.Type != entities.IPFS {
		return nil, errors.New("backnet should be of type IPFS")
	}

	env := []string{
		fmt.Sprintf("IPFS_PATH=%s", ib.ipfsDir),
	}

	errChan := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "ipfs", "daemon")
	cmd.Env = env

	ib.process = process{
		communityID: ib.communityID,
		processType: processTypeBacknet,
		context:     ctx,
		cancel:      cancel,
		errChan:     errChan,
	}

	// Start ipfs daemon
	go func(cmd *exec.Cmd, errChan chan error) {
		//cmd.Stdout = os.Stdout
		//cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			errChan <- err
		}
		err = cmd.Wait()
		if err != nil {
			errChan <- err
		}
	}(cmd, errChan)

	return &ib.process, nil
}
